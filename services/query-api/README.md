# query-api

The query API exposes read-only views backed by PostgreSQL projections.

## Endpoints

- `GET /health`
- `GET /ready`
- `GET /live`
- `GET /metrics`
- `GET /api/v1/customers`
- `GET /api/v1/customers/{customerID}`
- `GET /api/v1/orders`
- `GET /api/v1/orders/{orderID}`
- `GET /api/v1/invoices`
- `GET /api/v1/invoices/{invoiceID}`

## Features

- pagination via `limit` and `offset`
- base filters for customers, orders and invoices
- deterministic ordering for portfolio demos and smoke tests
