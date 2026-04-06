# sap-mock-api

The mock SAP service provides realistic upstream business payloads for integration testing and demos.

## Endpoints

- `GET /health`
- `GET /ready`
- `GET /live`
- `GET /metrics`
- `GET /api/v1/sample-data`
- `POST /api/v1/simulations/sales-orders/create`
- `POST /api/v1/simulations/sales-orders/update`
- `POST /api/v1/simulations/customers/update`
- `POST /api/v1/simulations/invoices/issue`

## Notes

- Request bodies are optional. When omitted, embedded sample payloads are used.
- Add `?dispatch=true` to forward the generated payload to `ingestion-api`.
- `AUTO_DISPATCH=true` enables dispatch by default.
