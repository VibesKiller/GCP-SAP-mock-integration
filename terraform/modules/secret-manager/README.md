# Secret Manager Module

Creates named Secret Manager secrets and, when requested, initial secret versions.

## Features

- generic map-driven secret creation
- optional initial secret value creation controlled by a static `create_initial_version` flag
- optional secret-level accessor bindings for least privilege
- accessor bindings are passed as a separate static-key map so service account emails can be computed during apply without breaking Terraform planning

This module is used to store generated application credentials and future integration secrets without hardcoding values in the repository.
