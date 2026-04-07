# Managed Kafka Module

Creates a Google Managed Service for Apache Kafka cluster together with platform topics and ACLs.

## What Is Managed Here

- managed Kafka cluster attached to the application VPC subnet
- explicit topic creation from repository-owned contracts
- ACLs derived from service ownership and consumer-group ownership

## Intentional Scope

This module provisions the Kafka control plane pieces that are practical in a platform demo.

It does not attempt to provision client-side bootstrap configuration into workloads automatically. Instead, it outputs the exact `gcloud` command required to retrieve the managed bootstrap address after apply.
