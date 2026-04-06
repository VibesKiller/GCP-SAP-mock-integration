package domain

import (
	"testing"
	"time"
)

func TestEventEnvelopeValidateSuccess(t *testing.T) {
	envelope := EventEnvelope{
		EventID:       "evt-123",
		EventType:     EventTypeSalesOrderCreated,
		Version:       EventVersionV1,
		Source:        SourceSAPS4HANA,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: "corr-123",
		Payload:       []byte(`{"sales_order_id":"SO-1"}`),
	}

	if err := envelope.Validate(); err != nil {
		t.Fatalf("expected valid envelope, got error: %v", err)
	}
}

func TestEventEnvelopeValidateUnsupportedEventType(t *testing.T) {
	envelope := EventEnvelope{
		EventID:       "evt-123",
		EventType:     "unknown.event",
		Version:       EventVersionV1,
		Source:        SourceSAPS4HANA,
		OccurredAt:    time.Now().UTC(),
		CorrelationID: "corr-123",
		Payload:       []byte(`{"foo":"bar"}`),
	}

	if err := envelope.Validate(); err == nil {
		t.Fatal("expected validation error for unsupported event type")
	}
}
