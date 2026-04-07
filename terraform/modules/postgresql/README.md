# PostgreSQL Module

Creates a Cloud SQL for PostgreSQL instance, an application database and an application user.

## Defaults

- PostgreSQL 16
- query insights enabled
- backups and point-in-time recovery enabled
- private IP preferred
- maintenance window defined explicitly
- configurable database and user deletion policies for clean destroy flows after application-owned objects exist

## Security Note

The module expects the application password as input. In the environment stacks, that password is generated with Terraform and stored in Secret Manager.
