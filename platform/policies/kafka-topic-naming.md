# Kafka Topic Naming Policy

Kafka topics are governed integration contracts and must remain stable over time.

## Convention

Primary business topics follow:

`<business-domain>.<entity-plural>.v<version>`

Platform topics follow:

`<business-domain>.integration.<purpose>.v<version>`

Examples:

- `sap.sales-orders.v1`
- `sap.customers.v1`
- `sap.invoices.v1`
- `sap.integration.dlq.v1`

## Rules

- Use lowercase topic names with dot-separated semantic segments.
- Put versioning in the final segment only.
- Keep event type in the payload or headers, not in the topic name.
- Use message keys that preserve aggregate ordering for the core entity.
- Document every consumer group in `platform/kafka/consumer-groups.yaml`.
- Add failure metadata headers for DLQ messages.
