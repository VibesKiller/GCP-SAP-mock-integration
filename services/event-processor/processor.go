package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kafkaGo "github.com/segmentio/kafka-go"

	"gcp-sap-mock-integration/internal/domain"
	platformHttp "gcp-sap-mock-integration/internal/platform/httpx"
	platformKafka "gcp-sap-mock-integration/internal/platform/kafka"
	platformPostgres "gcp-sap-mock-integration/internal/platform/postgres"
)

var errDuplicateEvent = errors.New("duplicate event")

type permanentError struct {
	err error
}

func (e permanentError) Error() string {
	return e.err.Error()
}

func (e permanentError) Unwrap() error {
	return e.err
}

type dlqMessage struct {
	FailureReason     string               `json:"failure_reason"`
	ErrorMessage      string               `json:"error_message"`
	FailedAt          time.Time            `json:"failed_at"`
	OriginalTopic     string               `json:"original_topic"`
	OriginalPartition int                  `json:"original_partition"`
	OriginalOffset    int64                `json:"original_offset"`
	OriginalKey       string               `json:"original_key"`
	Headers           map[string]string    `json:"headers"`
	Envelope          domain.EventEnvelope `json:"envelope"`
}

type app struct {
	config    appConfig
	logger    *slog.Logger
	db        *pgxpool.Pool
	dialer    *kafkaGo.Dialer
	reader    *kafkaGo.Reader
	dlqWriter *kafkaGo.Writer
	metrics   *metrics
}

func newApp(ctx context.Context, cfg appConfig, logger *slog.Logger) (*app, error) {
	db, err := platformPostgres.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		return nil, err
	}

	dialer, err := platformKafka.NewDialer(cfg.Kafka)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("build Kafka dialer: %w", err)
	}

	transport, err := platformKafka.NewTransport(cfg.Kafka)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("build Kafka transport: %w", err)
	}

	reader := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		GroupID:        cfg.KafkaConsumerGroup,
		GroupTopics:    cfg.KafkaTopics,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
		StartOffset:    kafkaGo.FirstOffset,
		Dialer:         dialer,
	})

	dlqWriter := &kafkaGo.Writer{
		Addr:                   kafkaGo.TCP(cfg.Kafka.Brokers...),
		Balancer:               &kafkaGo.Hash{},
		RequiredAcks:           kafkaGo.RequireAll,
		AllowAutoTopicCreation: false,
		BatchTimeout:           100 * time.Millisecond,
		Transport:              transport,
	}

	return &app{
		config:    cfg,
		logger:    logger,
		db:        db,
		dialer:    dialer,
		reader:    reader,
		dlqWriter: dlqWriter,
		metrics:   newMetrics(),
	}, nil
}

func (a *app) close() {
	a.reader.Close()
	a.dlqWriter.Close()
	a.db.Close()
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	platformHttp.RegisterHealthEndpoints(mux, a.config.ServiceName, a.ready)
	mux.Handle("/metrics", promhttp.Handler())
	return platformHttp.Chain(mux,
		platformHttp.CorrelationMiddleware(),
		platformHttp.RecoveryMiddleware(a.logger),
		platformHttp.LoggingMiddleware(a.logger),
	)
}

func (a *app) ready(ctx context.Context) error {
	if err := a.db.Ping(ctx); err != nil {
		return err
	}

	conn, err := a.dialer.DialContext(ctx, "tcp", a.config.Kafka.Brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
}

func (a *app) run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		message, err := a.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			a.logger.Error("fetch kafka message", "error", err)
			time.Sleep(a.config.RetryBackoff)
			continue
		}

		started := time.Now()
		envelope, outcome, err := a.processWithRetry(ctx, message)
		if errors.Is(err, errDuplicateEvent) {
			a.metrics.duplicatesTotal.Inc()
			a.metrics.consumedTotal.WithLabelValues(message.Topic, "duplicate").Inc()
			if commitErr := a.commitMessage(ctx, message); commitErr != nil {
				a.logger.Error("commit duplicate message", "error", commitErr, "topic", message.Topic, "offset", message.Offset)
			}
			continue
		}

		if err != nil {
			a.metrics.consumedTotal.WithLabelValues(message.Topic, "failed").Inc()
			a.logger.Error("message processing failed",
				"topic", message.Topic,
				"partition", message.Partition,
				"offset", message.Offset,
				"event_id", envelope.EventID,
				"event_type", envelope.EventType,
				"correlation_id", envelope.CorrelationID,
				"error", err,
			)
			time.Sleep(a.config.RetryBackoff)
			continue
		}

		if outcome == "dlq" {
			a.metrics.consumedTotal.WithLabelValues(message.Topic, "dlq").Inc()
		} else {
			a.metrics.consumedTotal.WithLabelValues(message.Topic, "processed").Inc()
			a.metrics.processingLatency.WithLabelValues(envelope.EventType).Observe(time.Since(started).Seconds())
		}

		if commitErr := a.commitMessage(ctx, message); commitErr != nil {
			a.logger.Error("commit processed message", "error", commitErr, "topic", message.Topic, "offset", message.Offset)
		}
	}
}

func (a *app) processWithRetry(ctx context.Context, message kafkaGo.Message) (domain.EventEnvelope, string, error) {
	var envelope domain.EventEnvelope
	if err := json.Unmarshal(message.Value, &envelope); err != nil {
		dlqErr := a.publishToDLQ(ctx, message, envelope, markPermanent(fmt.Errorf("decode event envelope: %w", err)))
		if dlqErr != nil {
			return domain.EventEnvelope{}, "", dlqErr
		}
		return domain.EventEnvelope{}, "dlq", nil
	}

	for attempt := 1; attempt <= a.config.RetryMaxAttempts; attempt++ {
		err := a.processMessage(ctx, message, envelope)
		switch {
		case err == nil:
			return envelope, "processed", nil
		case errors.Is(err, errDuplicateEvent):
			return envelope, "duplicate", err
		case isPermanent(err):
			if dlqErr := a.publishToDLQ(ctx, message, envelope, err); dlqErr != nil {
				return envelope, "", dlqErr
			}
			return envelope, "dlq", nil
		case isTransient(err) && attempt < a.config.RetryMaxAttempts:
			a.metrics.retriesTotal.WithLabelValues(envelope.EventType).Inc()
			a.logger.Warn("transient processing error, retrying",
				"attempt", attempt,
				"event_id", envelope.EventID,
				"event_type", envelope.EventType,
				"correlation_id", envelope.CorrelationID,
				"error", err,
			)
			time.Sleep(a.config.RetryBackoff)
		default:
			if dlqErr := a.publishToDLQ(ctx, message, envelope, err); dlqErr != nil {
				return envelope, "", dlqErr
			}
			return envelope, "dlq", nil
		}
	}

	return envelope, "", nil
}

func (a *app) processMessage(ctx context.Context, message kafkaGo.Message, envelope domain.EventEnvelope) error {
	if err := envelope.Validate(); err != nil {
		return markPermanent(fmt.Errorf("validate event envelope: %w", err))
	}

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	inserted, err := a.insertProcessedEvent(ctx, tx, message, envelope)
	if err != nil {
		return err
	}
	if !inserted {
		return errDuplicateEvent
	}

	if err := a.applyEvent(ctx, tx, envelope); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	a.logger.Info("event processed",
		"event_id", envelope.EventID,
		"event_type", envelope.EventType,
		"correlation_id", envelope.CorrelationID,
	)

	return nil
}

func (a *app) insertProcessedEvent(ctx context.Context, tx pgx.Tx, message kafkaGo.Message, envelope domain.EventEnvelope) (bool, error) {
	headersJSON, err := json.Marshal(headersToMap(message.Headers))
	if err != nil {
		return false, markPermanent(fmt.Errorf("encode kafka headers: %w", err))
	}

	commandTag, err := tx.Exec(ctx, `
    INSERT INTO processed_events (
      event_id,
      event_type,
      version,
      source,
      occurred_at,
      correlation_id,
      kafka_topic,
      kafka_partition,
      kafka_offset,
      kafka_key,
      kafka_headers,
      payload,
      status
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
    ON CONFLICT (event_id) DO NOTHING
  `,
		envelope.EventID,
		envelope.EventType,
		envelope.Version,
		envelope.Source,
		envelope.OccurredAt,
		envelope.CorrelationID,
		message.Topic,
		message.Partition,
		message.Offset,
		string(message.Key),
		headersJSON,
		envelope.Payload,
		"processed",
	)
	if err != nil {
		return false, err
	}

	return commandTag.RowsAffected() > 0, nil
}

func (a *app) applyEvent(ctx context.Context, tx pgx.Tx, envelope domain.EventEnvelope) error {
	switch envelope.EventType {
	case domain.EventTypeSalesOrderCreated, domain.EventTypeSalesOrderUpdated:
		var payload domain.SalesOrderPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return markPermanent(fmt.Errorf("decode sales order payload: %w", err))
		}
		return a.applySalesOrder(ctx, tx, envelope, payload)
	case domain.EventTypeCustomerUpdated:
		var payload domain.CustomerPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return markPermanent(fmt.Errorf("decode customer payload: %w", err))
		}
		return a.applyCustomer(ctx, tx, envelope, payload)
	case domain.EventTypeInvoiceIssued:
		var payload domain.InvoicePayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return markPermanent(fmt.Errorf("decode invoice payload: %w", err))
		}
		return a.applyInvoice(ctx, tx, envelope, payload)
	default:
		return markPermanent(fmt.Errorf("unsupported event_type %q", envelope.EventType))
	}
}

func (a *app) applyCustomer(ctx context.Context, tx pgx.Tx, envelope domain.EventEnvelope, payload domain.CustomerPayload) error {
	_, err := tx.Exec(ctx, `
    INSERT INTO customers (
      customer_id, customer_number, full_name, email, phone, country_code, city, postal_code, segment, status,
      last_event_id, last_correlation_id, source_updated_at, created_at, updated_at
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
    ON CONFLICT (customer_id) DO UPDATE SET
      customer_number = EXCLUDED.customer_number,
      full_name = EXCLUDED.full_name,
      email = EXCLUDED.email,
      phone = EXCLUDED.phone,
      country_code = EXCLUDED.country_code,
      city = EXCLUDED.city,
      postal_code = EXCLUDED.postal_code,
      segment = EXCLUDED.segment,
      status = EXCLUDED.status,
      last_event_id = EXCLUDED.last_event_id,
      last_correlation_id = EXCLUDED.last_correlation_id,
      source_updated_at = EXCLUDED.source_updated_at,
      updated_at = NOW()
  `,
		payload.CustomerID,
		payload.CustomerNumber,
		payload.FullName,
		payload.Email,
		payload.Phone,
		payload.CountryCode,
		payload.City,
		payload.PostalCode,
		payload.Segment,
		payload.Status,
		envelope.EventID,
		envelope.CorrelationID,
		envelope.OccurredAt,
	)
	return err
}

func (a *app) applySalesOrder(ctx context.Context, tx pgx.Tx, envelope domain.EventEnvelope, payload domain.SalesOrderPayload) error {
	requestedDeliveryDate, err := parseDateOnly(payload.RequestedDeliveryDate)
	if err != nil {
		return markPermanent(fmt.Errorf("parse requested_delivery_date: %w", err))
	}
	documentDate, err := parseDateOnly(payload.DocumentDate)
	if err != nil {
		return markPermanent(fmt.Errorf("parse document_date: %w", err))
	}

	if err := a.ensureCustomerStub(ctx, tx, payload.CustomerID, envelope); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
    INSERT INTO orders (
      order_id, customer_id, sales_org, distribution_channel, division, currency, status,
      requested_delivery_date, document_date, net_amount, tax_amount, total_amount,
      last_event_id, last_correlation_id, source_updated_at, created_at, updated_at
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW(),NOW())
    ON CONFLICT (order_id) DO UPDATE SET
      customer_id = EXCLUDED.customer_id,
      sales_org = EXCLUDED.sales_org,
      distribution_channel = EXCLUDED.distribution_channel,
      division = EXCLUDED.division,
      currency = EXCLUDED.currency,
      status = EXCLUDED.status,
      requested_delivery_date = EXCLUDED.requested_delivery_date,
      document_date = EXCLUDED.document_date,
      net_amount = EXCLUDED.net_amount,
      tax_amount = EXCLUDED.tax_amount,
      total_amount = EXCLUDED.total_amount,
      last_event_id = EXCLUDED.last_event_id,
      last_correlation_id = EXCLUDED.last_correlation_id,
      source_updated_at = EXCLUDED.source_updated_at,
      updated_at = NOW()
  `,
		payload.SalesOrderID,
		payload.CustomerID,
		payload.SalesOrg,
		payload.DistributionChannel,
		payload.Division,
		payload.Currency,
		payload.Status,
		requestedDeliveryDate,
		documentDate,
		payload.Totals.NetAmount,
		payload.Totals.TaxAmount,
		payload.Totals.TotalAmount,
		envelope.EventID,
		envelope.CorrelationID,
		envelope.OccurredAt,
	)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, payload.SalesOrderID); err != nil {
		return err
	}

	for _, item := range payload.Items {
		_, err := tx.Exec(ctx, `
      INSERT INTO order_items (
        order_id, line_number, sku, description, quantity, unit, unit_price, net_amount, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())
    `,
			payload.SalesOrderID,
			item.LineNumber,
			item.SKU,
			item.Description,
			item.Quantity,
			item.Unit,
			item.UnitPrice,
			item.NetAmount,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *app) applyInvoice(ctx context.Context, tx pgx.Tx, envelope domain.EventEnvelope, payload domain.InvoicePayload) error {
	issueDate, err := parseDateOrTime(payload.IssueDate)
	if err != nil {
		return markPermanent(fmt.Errorf("parse issue_date: %w", err))
	}
	dueDate, err := parseDateOrTime(payload.DueDate)
	if err != nil {
		return markPermanent(fmt.Errorf("parse due_date: %w", err))
	}

	if err := a.ensureCustomerStub(ctx, tx, payload.CustomerID, envelope); err != nil {
		return err
	}
	if err := a.ensureOrderStub(ctx, tx, payload, envelope, issueDate); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
    INSERT INTO invoices (
      invoice_id, order_id, customer_id, currency, status, issue_date, due_date,
      net_amount, tax_amount, total_amount, last_event_id, last_correlation_id, source_updated_at,
      created_at, updated_at
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
    ON CONFLICT (invoice_id) DO UPDATE SET
      order_id = EXCLUDED.order_id,
      customer_id = EXCLUDED.customer_id,
      currency = EXCLUDED.currency,
      status = EXCLUDED.status,
      issue_date = EXCLUDED.issue_date,
      due_date = EXCLUDED.due_date,
      net_amount = EXCLUDED.net_amount,
      tax_amount = EXCLUDED.tax_amount,
      total_amount = EXCLUDED.total_amount,
      last_event_id = EXCLUDED.last_event_id,
      last_correlation_id = EXCLUDED.last_correlation_id,
      source_updated_at = EXCLUDED.source_updated_at,
      updated_at = NOW()
  `,
		payload.InvoiceID,
		payload.SalesOrderID,
		payload.CustomerID,
		payload.Currency,
		payload.Status,
		issueDate,
		dueDate,
		payload.NetAmount,
		payload.TaxAmount,
		payload.TotalAmount,
		envelope.EventID,
		envelope.CorrelationID,
		envelope.OccurredAt,
	)
	return err
}

func (a *app) ensureCustomerStub(ctx context.Context, tx pgx.Tx, customerID string, envelope domain.EventEnvelope) error {
	_, err := tx.Exec(ctx, `
    INSERT INTO customers (
      customer_id, customer_number, full_name, email, phone, country_code, city, postal_code, segment, status,
      last_event_id, last_correlation_id, source_updated_at, created_at, updated_at
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
    ON CONFLICT (customer_id) DO NOTHING
  `,
		customerID,
		customerID,
		fmt.Sprintf("Pending master data for %s", customerID),
		fmt.Sprintf("unknown+%s@example.invalid", strings.ToLower(customerID)),
		"",
		"ZZ",
		"",
		"",
		"unknown",
		"PENDING_MASTER_DATA",
		envelope.EventID,
		envelope.CorrelationID,
		envelope.OccurredAt,
	)
	return err
}

func (a *app) ensureOrderStub(ctx context.Context, tx pgx.Tx, invoice domain.InvoicePayload, envelope domain.EventEnvelope, issueDate time.Time) error {
	_, err := tx.Exec(ctx, `
    INSERT INTO orders (
      order_id, customer_id, sales_org, distribution_channel, division, currency, status,
      requested_delivery_date, document_date, net_amount, tax_amount, total_amount,
      last_event_id, last_correlation_id, source_updated_at, created_at, updated_at
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW(),NOW())
    ON CONFLICT (order_id) DO NOTHING
  `,
		invoice.SalesOrderID,
		invoice.CustomerID,
		"UNKNOWN",
		"UNKNOWN",
		"UNKNOWN",
		invoice.Currency,
		"PENDING_ORDER_EVENT",
		issueDate,
		issueDate,
		invoice.NetAmount,
		invoice.TaxAmount,
		invoice.TotalAmount,
		envelope.EventID,
		envelope.CorrelationID,
		envelope.OccurredAt,
	)
	return err
}

func (a *app) publishToDLQ(ctx context.Context, message kafkaGo.Message, envelope domain.EventEnvelope, failure error) error {
	reason := failureReason(failure)
	originalKey := string(message.Key)
	if originalKey == "" {
		originalKey = envelope.EventID
	}

	payload := dlqMessage{
		FailureReason:     reason,
		ErrorMessage:      failure.Error(),
		FailedAt:          time.Now().UTC(),
		OriginalTopic:     message.Topic,
		OriginalPartition: message.Partition,
		OriginalOffset:    message.Offset,
		OriginalKey:       string(message.Key),
		Headers:           headersToMap(message.Headers),
		Envelope:          envelope,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return markPermanent(fmt.Errorf("marshal dlq payload: %w", err))
	}

	dlqHeaders := []kafkaGo.Header{
		{Key: "failure_reason", Value: []byte(reason)},
		{Key: "original_topic", Value: []byte(message.Topic)},
		{Key: "original_partition", Value: []byte(fmt.Sprintf("%d", message.Partition))},
		{Key: "original_offset", Value: []byte(fmt.Sprintf("%d", message.Offset))},
		{Key: "original_key", Value: []byte(originalKey)},
		{Key: "partition_key", Value: []byte(originalKey)},
	}

	if envelope.EventID != "" {
		dlqHeaders = append(dlqHeaders,
			kafkaGo.Header{Key: "event_id", Value: []byte(envelope.EventID)},
			kafkaGo.Header{Key: "event_type", Value: []byte(envelope.EventType)},
			kafkaGo.Header{Key: "version", Value: []byte(envelope.Version)},
			kafkaGo.Header{Key: "source", Value: []byte(envelope.Source)},
			kafkaGo.Header{Key: "correlation_id", Value: []byte(envelope.CorrelationID)},
		)
	}

	if err := a.dlqWriter.WriteMessages(ctx, kafkaGo.Message{
		Topic:   a.config.KafkaDLQTopic,
		Key:     []byte(originalKey),
		Value:   payloadBytes,
		Time:    time.Now().UTC(),
		Headers: dlqHeaders,
	}); err != nil {
		return fmt.Errorf("publish message to dlq: %w", err)
	}

	a.metrics.dlqTotal.WithLabelValues(reason).Inc()
	a.logger.Warn("message published to dlq",
		"reason", reason,
		"event_id", envelope.EventID,
		"event_type", envelope.EventType,
		"correlation_id", envelope.CorrelationID,
		"original_topic", message.Topic,
		"offset", message.Offset,
	)

	if err := a.recordDLQEvent(ctx, message, envelope, reason); err != nil {
		return fmt.Errorf("record dlq event metadata: %w", err)
	}

	return nil
}

func (a *app) recordDLQEvent(ctx context.Context, message kafkaGo.Message, envelope domain.EventEnvelope, status string) error {
	if envelope.EventID == "" {
		return nil
	}

	headersJSON, err := json.Marshal(headersToMap(message.Headers))
	if err != nil {
		return err
	}

	_, err = a.db.Exec(ctx, `
    INSERT INTO processed_events (
      event_id,
      event_type,
      version,
      source,
      occurred_at,
      correlation_id,
      kafka_topic,
      kafka_partition,
      kafka_offset,
      kafka_key,
      kafka_headers,
      payload,
      status,
      processed_at
    ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW())
    ON CONFLICT (event_id) DO UPDATE SET
      kafka_topic = EXCLUDED.kafka_topic,
      kafka_partition = EXCLUDED.kafka_partition,
      kafka_offset = EXCLUDED.kafka_offset,
      kafka_key = EXCLUDED.kafka_key,
      kafka_headers = EXCLUDED.kafka_headers,
      payload = EXCLUDED.payload,
      status = EXCLUDED.status,
      processed_at = NOW()
  `,
		envelope.EventID,
		envelope.EventType,
		envelope.Version,
		envelope.Source,
		envelope.OccurredAt,
		envelope.CorrelationID,
		message.Topic,
		message.Partition,
		message.Offset,
		string(message.Key),
		headersJSON,
		envelope.Payload,
		status,
	)
	return err
}

func (a *app) commitMessage(ctx context.Context, message kafkaGo.Message) error {
	commitCtx, cancel := context.WithTimeout(ctx, a.config.KafkaCommitTimeout)
	defer cancel()
	return a.reader.CommitMessages(commitCtx, message)
}

func headersToMap(headers []kafkaGo.Header) map[string]string {
	output := make(map[string]string, len(headers))
	for _, header := range headers {
		output[header.Key] = string(header.Value)
	}
	return output
}

func parseDateOnly(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func parseDateOrTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format %q", value)
}

func markPermanent(err error) error {
	if err == nil {
		return nil
	}
	return permanentError{err: err}
}

func isPermanent(err error) bool {
	var target permanentError
	return errors.As(err, &target)
}

func isTransient(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		code := pgErr.Code
		return strings.HasPrefix(code, "08") || strings.HasPrefix(code, "53") || code == "40001" || code == "57P01"
	}

	return errors.Is(err, context.DeadlineExceeded)
}

func failureReason(err error) string {
	switch {
	case isPermanent(err):
		return "non_recoverable"
	case isTransient(err):
		return "transient_retries_exhausted"
	default:
		return "processing_error"
	}
}
