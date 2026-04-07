package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kafkaGo "github.com/segmentio/kafka-go"

	"gcp-sap-mock-integration/internal/domain"
	"gcp-sap-mock-integration/internal/platform/httpx"
	platformKafka "gcp-sap-mock-integration/internal/platform/kafka"
)

type publishMetadata struct {
	Topic         string `json:"topic"`
	MessageKey    string `json:"message_key"`
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	CorrelationID string `json:"correlation_id"`
	OccurredAt    string `json:"occurred_at"`
	Version       string `json:"version"`
}

type app struct {
	config  appConfig
	logger  *slog.Logger
	dialer  *kafkaGo.Dialer
	writer  *kafkaGo.Writer
	metrics *metrics
}

func newApp(cfg appConfig, logger *slog.Logger) (*app, error) {
	transport, err := platformKafka.NewTransport(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("build Kafka transport: %w", err)
	}

	dialer, err := platformKafka.NewDialer(cfg.Kafka)
	if err != nil {
		return nil, fmt.Errorf("build Kafka dialer: %w", err)
	}

	writer := &kafkaGo.Writer{
		Addr:                   kafkaGo.TCP(cfg.Kafka.Brokers...),
		Balancer:               &kafkaGo.Hash{},
		RequiredAcks:           kafkaGo.RequireAll,
		AllowAutoTopicCreation: false,
		WriteTimeout:           cfg.KafkaWriteTimeout,
		Async:                  false,
		BatchTimeout:           100 * time.Millisecond,
		Transport:              transport,
	}

	return &app{
		config:  cfg,
		logger:  logger,
		dialer:  dialer,
		writer:  writer,
		metrics: newMetrics(),
	}, nil
}

func (a *app) close() error {
	return a.writer.Close()
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	httpx.RegisterHealthEndpoints(mux, a.config.ServiceName, a.ready)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("POST /api/v1/sap/sales-orders", a.handleSalesOrderCreated)
	mux.HandleFunc("PATCH /api/v1/sap/sales-orders/{orderID}", a.handleSalesOrderUpdated)
	mux.HandleFunc("PATCH /api/v1/sap/customers/{customerID}", a.handleCustomerUpdated)
	mux.HandleFunc("POST /api/v1/sap/invoices", a.handleInvoiceIssued)

	return httpx.Chain(mux,
		httpx.CorrelationMiddleware(),
		httpx.RecoveryMiddleware(a.logger),
		httpx.LoggingMiddleware(a.logger),
	)
}

func (a *app) ready(ctx context.Context) error {
	conn, err := a.dialer.DialContext(ctx, "tcp", a.config.Kafka.Brokers[0])
	if err != nil {
		return err
	}
	return conn.Close()
}

func (a *app) handleSalesOrderCreated(w http.ResponseWriter, r *http.Request) {
	var payload domain.SAPSalesOrderPayload
	if err := httpx.DecodeJSON(r, &payload); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderCreated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	if err := payload.Validate(); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderCreated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	metadata, err := a.publishSalesOrder(r.Context(), domain.EventTypeSalesOrderCreated, payload)
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderCreated, "error").Inc()
		httpx.WriteError(w, http.StatusBadGateway, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderCreated, "accepted").Inc()
	httpx.WriteJSON(w, http.StatusAccepted, metadata)
}

func (a *app) handleSalesOrderUpdated(w http.ResponseWriter, r *http.Request) {
	var payload domain.SAPSalesOrderPayload
	if err := httpx.DecodeJSON(r, &payload); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	if err := payload.Validate(); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	if pathID := r.PathValue("orderID"); pathID != "" && pathID != payload.SalesDocumentID {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, "path orderID must match sales_document_id", httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	metadata, err := a.publishSalesOrder(r.Context(), domain.EventTypeSalesOrderUpdated, payload)
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderUpdated, "error").Inc()
		httpx.WriteError(w, http.StatusBadGateway, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues(domain.EventTypeSalesOrderUpdated, "accepted").Inc()
	httpx.WriteJSON(w, http.StatusAccepted, metadata)
}

func (a *app) handleCustomerUpdated(w http.ResponseWriter, r *http.Request) {
	var payload domain.SAPCustomerPayload
	if err := httpx.DecodeJSON(r, &payload); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeCustomerUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	if err := payload.Validate(); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeCustomerUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	if pathID := r.PathValue("customerID"); pathID != "" && pathID != payload.CustomerID {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeCustomerUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, "path customerID must match customer_id", httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	metadata, err := a.publishCustomer(r.Context(), payload)
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeCustomerUpdated, "error").Inc()
		httpx.WriteError(w, http.StatusBadGateway, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues(domain.EventTypeCustomerUpdated, "accepted").Inc()
	httpx.WriteJSON(w, http.StatusAccepted, metadata)
}

func (a *app) handleInvoiceIssued(w http.ResponseWriter, r *http.Request) {
	var payload domain.SAPInvoicePayload
	if err := httpx.DecodeJSON(r, &payload); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeInvoiceIssued, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	if err := payload.Validate(); err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeInvoiceIssued, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	metadata, err := a.publishInvoice(r.Context(), payload)
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues(domain.EventTypeInvoiceIssued, "error").Inc()
		httpx.WriteError(w, http.StatusBadGateway, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues(domain.EventTypeInvoiceIssued, "accepted").Inc()
	httpx.WriteJSON(w, http.StatusAccepted, metadata)
}

func (a *app) publishSalesOrder(ctx context.Context, eventType string, payload domain.SAPSalesOrderPayload) (publishMetadata, error) {
	canonical := domain.NormalizeSalesOrderPayload(payload)
	return a.publishEnvelope(ctx, eventType, canonical)
}

func (a *app) publishCustomer(ctx context.Context, payload domain.SAPCustomerPayload) (publishMetadata, error) {
	canonical := domain.NormalizeCustomerPayload(payload)
	return a.publishEnvelope(ctx, domain.EventTypeCustomerUpdated, canonical)
}

func (a *app) publishInvoice(ctx context.Context, payload domain.SAPInvoicePayload) (publishMetadata, error) {
	canonical := domain.NormalizeInvoicePayload(payload)
	return a.publishEnvelope(ctx, domain.EventTypeInvoiceIssued, canonical)
}

func (a *app) publishEnvelope(ctx context.Context, eventType string, payload any) (publishMetadata, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return publishMetadata{}, fmt.Errorf("marshal canonical payload: %w", err)
	}

	envelope := domain.EventEnvelope{
		EventID:       uuid.NewString(),
		EventType:     eventType,
		Version:       domain.EventVersionV1,
		Source:        a.config.SourceSystem,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: httpx.CorrelationIDFromContext(ctx),
		Payload:       payloadBytes,
	}

	if err := envelope.Validate(); err != nil {
		return publishMetadata{}, fmt.Errorf("validate event envelope: %w", err)
	}

	topic, err := platformKafka.TopicForEventType(eventType)
	if err != nil {
		return publishMetadata{}, err
	}

	messageKey, err := platformKafka.MessageKeyForEnvelope(envelope)
	if err != nil {
		return publishMetadata{}, err
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return publishMetadata{}, fmt.Errorf("marshal event envelope: %w", err)
	}

	started := time.Now()
	err = a.writer.WriteMessages(ctx, kafkaGo.Message{
		Topic:   topic,
		Key:     []byte(messageKey),
		Value:   envelopeBytes,
		Time:    envelope.OccurredAt,
		Headers: platformKafka.HeadersForEnvelope(envelope, messageKey),
	})
	a.metrics.publishDuration.WithLabelValues(topic, eventType).Observe(time.Since(started).Seconds())
	if err != nil {
		a.metrics.publishedTotal.WithLabelValues(topic, eventType, "error").Inc()
		return publishMetadata{}, fmt.Errorf("publish event to kafka: %w", err)
	}

	a.metrics.publishedTotal.WithLabelValues(topic, eventType, "success").Inc()
	a.logger.Info("event published to kafka",
		"topic", topic,
		"message_key", messageKey,
		"event_id", envelope.EventID,
		"event_type", envelope.EventType,
		"correlation_id", envelope.CorrelationID,
	)

	return publishMetadata{
		Topic:         topic,
		MessageKey:    messageKey,
		EventID:       envelope.EventID,
		EventType:     envelope.EventType,
		CorrelationID: envelope.CorrelationID,
		OccurredAt:    envelope.OccurredAt.Format(time.RFC3339),
		Version:       envelope.Version,
	}, nil
}
