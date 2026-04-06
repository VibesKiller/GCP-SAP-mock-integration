resource "google_sql_database_instance" "this" {
  name             = var.instance_name
  region           = var.region
  database_version = "POSTGRES_16"

  deletion_protection = var.deletion_protection

  settings {
    tier              = var.tier
    availability_type = var.availability_type
    disk_size         = var.disk_size_gb
    disk_autoresize   = true

    backup_configuration {
      enabled                        = true
      point_in_time_recovery_enabled = true
    }

    insights_config {
      query_insights_enabled = true
    }

    ip_configuration {
      ipv4_enabled    = var.ipv4_enabled
      private_network = var.private_network
    }
  }
}

resource "google_sql_database" "app" {
  name     = var.database_name
  instance = google_sql_database_instance.this.name
}

resource "google_sql_user" "app" {
  name     = var.username
  instance = google_sql_database_instance.this.name
  password = var.password
}
