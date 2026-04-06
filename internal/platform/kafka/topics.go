package kafka

import (
	"encoding/json"
	"fmt"
	"strings"

	kafkaGo "github.com/segmentio/kafka-go"

	"gcp-sap-mock-integration/internal/domain"
)

const (
	TopicSalesOrders            = "sap.sales-orders.v1"
	TopicCustomers              = "sap.customers.v1"
	TopicInvoices               = "sap.invoices.v1"
	TopicIntegrationDLQ         = "sap.integration.dlq.v1"
	ConsumerGroupEventProcessor = "sap-integration.event-processor.v1"
)

type TopicDefinition struct {
	Name              string   `json:"name" yaml:"name"`
	Owner             string   `json:"owner" yaml:"owner"`
	MessageKey        string   `json:"message_key" yaml:"messageKey"`
	PartitionStrategy string   `json:"partition_strategy" yaml:"partitionStrategy"`
	Headers           []string `json:"headers" yaml:"headers"`
	Description       string   `json:"description" yaml:"description"`
}

func TopicForEventType(eventType string) (string, error) {
	switch eventType {
	case domain.EventTypeSalesOrderCreated, domain.EventTypeSalesOrderUpdated:
		return TopicSalesOrders, nil
	case domain.EventTypeCustomerUpdated:
		return TopicCustomers, nil
	case domain.EventTypeInvoiceIssued:
		return TopicInvoices, nil
	default:
		return "", fmt.Errorf("unsupported event type %q", eventType)
	}
}

func MessageKeyForEnvelope(envelope domain.EventEnvelope) (string, error) {
	switch envelope.EventType {
	case domain.EventTypeSalesOrderCreated, domain.EventTypeSalesOrderUpdated:
		var payload domain.SalesOrderPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return "", fmt.Errorf("decode sales order payload: %w", err)
		}
		if strings.TrimSpace(payload.SalesOrderID) == "" {
			return "", fmt.Errorf("sales_order_id is required for key derivation")
		}
		return payload.SalesOrderID, nil
	case domain.EventTypeCustomerUpdated:
		var payload domain.CustomerPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return "", fmt.Errorf("decode customer payload: %w", err)
		}
		if strings.TrimSpace(payload.CustomerID) == "" {
			return "", fmt.Errorf("customer_id is required for key derivation")
		}
		return payload.CustomerID, nil
	case domain.EventTypeInvoiceIssued:
		var payload domain.InvoicePayload
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return "", fmt.Errorf("decode invoice payload: %w", err)
		}
		if strings.TrimSpace(payload.InvoiceID) == "" {
			return "", fmt.Errorf("invoice_id is required for key derivation")
		}
		return payload.InvoiceID, nil
	default:
		return "", fmt.Errorf("unsupported event type %q", envelope.EventType)
	}
}

func HeadersForEnvelope(envelope domain.EventEnvelope, messageKey string) []kafkaGo.Header {
	return []kafkaGo.Header{
		{Key: "event_id", Value: []byte(envelope.EventID)},
		{Key: "event_type", Value: []byte(envelope.EventType)},
		{Key: "version", Value: []byte(envelope.Version)},
		{Key: "source", Value: []byte(envelope.Source)},
		{Key: "correlation_id", Value: []byte(envelope.CorrelationID)},
		{Key: "partition_key", Value: []byte(messageKey)},
	}
}

func TopicCatalog() []TopicDefinition {
	commonHeaders := []string{"event_id", "event_type", "version", "source", "correlation_id", "partition_key"}
	return []TopicDefinition{
		{
			Name:              TopicSalesOrders,
			Owner:             "ingestion-api",
			MessageKey:        "sales_order_id",
			PartitionStrategy: "Hash partition by sales order to preserve aggregate ordering across create and update events.",
			Headers:           commonHeaders,
			Description:       "Canonical sales order events normalized from SAP REST payloads.",
		},
		{
			Name:              TopicCustomers,
			Owner:             "ingestion-api",
			MessageKey:        "customer_id",
			PartitionStrategy: "Hash partition by customer ID to keep master-data updates ordered.",
			Headers:           commonHeaders,
			Description:       "Canonical customer update events normalized from SAP REST payloads.",
		},
		{
			Name:              TopicInvoices,
			Owner:             "ingestion-api",
			MessageKey:        "invoice_id",
			PartitionStrategy: "Hash partition by invoice ID to preserve invoice lifecycle sequencing.",
			Headers:           commonHeaders,
			Description:       "Canonical invoice events normalized from SAP REST payloads.",
		},
		{
			Name:              TopicIntegrationDLQ,
			Owner:             "event-processor",
			MessageKey:        "original_partition_key",
			PartitionStrategy: "Reuse the original business key when available to simplify operational replay.",
			Headers: append(commonHeaders,
				"failure_reason",
				"original_topic",
				"original_partition",
				"original_offset",
			),
			Description: "Dead-letter topic for non-recoverable processor failures and exhausted retries.",
		},
	}
}
