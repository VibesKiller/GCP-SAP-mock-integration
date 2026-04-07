# Runtime Security Baseline

The initial runtime security posture for this project is intentionally pragmatic and interview-ready.

## Baseline Controls

- Containers run as non-root where the base image permits it.
- Images are expected to be immutable and versioned in Artifact Registry.
- Helm values expose resource requests and limits by default.
- Kubernetes probes are mandatory for every long-running service.
- Secrets are mounted or injected at runtime, never committed to Git.
- Structured JSON logs must include `service`, `environment` and `correlation_id` fields.
