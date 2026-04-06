# Terraform Prod Environment

The prod environment keeps the same module composition as dev while hardening runtime posture.

Characteristics:

- deletion protection enabled by default
- regional Cloud SQL high availability
- dedicated node service account
- stricter operational expectations around rollout, backup and change approval
