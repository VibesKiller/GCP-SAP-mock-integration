package kafka

import (
	"encoding/json"
	"testing"
	"time"

	"gcp-sap-mock-integration/internal/domain"
)

func TestTopicForEventType(t *testing.T) {
	tests := []struct {
		eventType string
		wantTopic string
	}{
		{eventType: domain.EventTypeSalesOrderCreated, wantTopic: TopicSalesOrders},
		{eventType: domain.EventTypeSalesOrderUpdated, wantTopic: TopicSalesOrders},
		{eventType: domain.EventTypeCustomerUpdated, wantTopic: TopicCustomers},
		{eventType: domain.EventTypeInvoiceIssued, wantTopic: TopicInvoices},
	}

	for _, tt := range tests {
		got, err := TopicForEventType(tt.eventType)
		if err != nil {
			t.Fatalf("expected topic for %s, got error: %v", tt.eventType, err)
		}
		if got != tt.wantTopic {
			t.Fatalf("expected topic %q for %s, got %q", tt.wantTopic, tt.eventType, got)
		}
	}
}

func TestMessageKeyForEnvelope(t *testing.T) {
	payloadBytes, err := json.Marshal(domain.SalesOrderPayload{SalesOrderID: "SO-2026-000184"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	envelope := domain.EventEnvelope{
		EventID:       "evt-1",
		EventType:     domain.EventTypeSalesOrderCreated,
		Version:       domain.EventVersionV1,
		Source:        domain.SourceSAPS4HANA,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: "corr-1",
		Payload:       payloadBytes,
	}

	key, err := MessageKeyForEnvelope(envelope)
	if err != nil {
		t.Fatalf("expected key, got error: %v", err)
	}
	if key != "SO-2026-000184" {
		t.Fatalf("expected message key SO-2026-000184, got %q", key)
	}
}

func TestHeadersForEnvelope(t *testing.T) {
	envelope := domain.EventEnvelope{
		EventID:       "evt-1",
		EventType:     domain.EventTypeCustomerUpdated,
		Version:       domain.EventVersionV1,
		Source:        domain.SourceSAPS4HANA,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: "corr-1",
		Payload:       []byte(`{"customer_id":"CUST-1"}`),
	}

	headers := HeadersForEnvelope(envelope, "CUST-1")
	if len(headers) != 6 {
		t.Fatalf("expected 6 headers, got %d", len(headers))
	}
	if string(headers[4].Value) != envelope.CorrelationID {
		t.Fatalf("expected correlation_id header %q, got %q", envelope.CorrelationID, string(headers[4].Value))
	}
}
