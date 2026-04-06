# event-processor

The processor consumes canonical Kafka events, applies idempotency checks, persists projections in PostgreSQL and routes non-recoverable messages to the DLQ.

## Runtime Endpoints

- `GET /health`
- `GET /ready`
- `GET /live`
- `GET /metrics`

## Responsibilities

- consume `sap.sales-orders.v1`, `sap.customers.v1` and `sap.invoices.v1`
- write `processed_events` for auditability and deduplication
- persist customers, orders, order items and invoices
- retry transient failures with bounded backoff
- publish poison messages to `sap.integration.dlq.v1`
