package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	EventVersionV1             = "v1"
	SourceSAPS4HANA            = "sap-s4hana"
	EventTypeSalesOrderCreated = "sales_order.created"
	EventTypeSalesOrderUpdated = "sales_order.updated"
	EventTypeCustomerUpdated   = "customer.updated"
	EventTypeInvoiceIssued     = "invoice.issued"
)

type EventEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	Version       string          `json:"version"`
	Source        string          `json:"source"`
	OccurredAt    time.Time       `json:"occurred_at"`
	CorrelationID string          `json:"correlation_id"`
	Payload       json.RawMessage `json:"payload"`
}

func (e EventEnvelope) Validate() error {
	switch {
	case strings.TrimSpace(e.EventID) == "":
		return errors.New("event_id is required")
	case strings.TrimSpace(e.EventType) == "":
		return errors.New("event_type is required")
	case strings.TrimSpace(e.Version) == "":
		return errors.New("version is required")
	case strings.TrimSpace(e.Source) == "":
		return errors.New("source is required")
	case e.OccurredAt.IsZero():
		return errors.New("occurred_at is required")
	case strings.TrimSpace(e.CorrelationID) == "":
		return errors.New("correlation_id is required")
	case len(e.Payload) == 0:
		return errors.New("payload is required")
	}

	switch e.EventType {
	case EventTypeSalesOrderCreated, EventTypeSalesOrderUpdated, EventTypeCustomerUpdated, EventTypeInvoiceIssued:
		return nil
	default:
		return fmt.Errorf("unsupported event_type %q", e.EventType)
	}
}
