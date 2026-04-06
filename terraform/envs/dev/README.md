# Terraform Dev Environment

The dev environment intentionally optimizes for fast iteration while preserving the same architectural boundaries as production.

Characteristics:

- single GKE cluster
- single Artifact Registry repository
- single Cloud SQL PostgreSQL instance
- baseline secret placeholders in Secret Manager
- lower protection settings to keep iterative development practical
