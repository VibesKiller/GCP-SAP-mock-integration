# Network Module

Creates the baseline VPC used by GKE, Cloud SQL private IP and Managed Kafka.

## Resources

- custom-mode VPC
- primary subnet with GKE secondary ranges
- subnet flow logs
- Cloud Router and Cloud NAT
- private service access for Google managed services

## Notes

- Cloud NAT is enabled by default because the GKE module uses private nodes.
- Private service access is enabled by default so Cloud SQL can use private IP.
- Set `private_service_range_address` explicitly when the platform has a deterministic IP plan. This avoids GCP auto-selecting a range that overlaps GKE secondary ranges.
- `private_service_access_deletion_policy` can be set to `ABANDON` for dev teardown workflows where Service Networking keeps a deleted producer service attached for a short period.
