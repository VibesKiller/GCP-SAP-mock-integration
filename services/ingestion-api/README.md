# ingestion-api

The ingestion service is the authenticated integration edge between SAP-style REST payloads and the Kafka event backbone.

## Endpoints

- `GET /health`
- `GET /ready`
- `GET /live`
- `GET /metrics`
- `POST /api/v1/sap/sales-orders`
- `PATCH /api/v1/sap/sales-orders/{orderID}`
- `PATCH /api/v1/sap/customers/{customerID}`
- `POST /api/v1/sap/invoices`

## Responsibilities

- validate incoming SAP payloads
- normalize them into the canonical event envelope
- enrich events with metadata such as correlation ID and event ID
- publish to Kafka with stable topic and key strategy
