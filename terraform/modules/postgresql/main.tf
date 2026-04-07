resource "google_sql_database_instance" "this" {
  name             = var.instance_name
  region           = var.region
  database_version = var.database_version

  deletion_protection = var.deletion_protection

  settings {
    tier              = var.tier
    edition           = var.edition
    availability_type = var.availability_type
    disk_type         = var.disk_type
    disk_size         = var.disk_size_gb
    disk_autoresize   = true
    user_labels       = var.labels

    backup_configuration {
      enabled                        = var.backup_enabled
      point_in_time_recovery_enabled = var.point_in_time_recovery_enabled
    }

    insights_config {
      query_insights_enabled  = true
      query_string_length     = 2048
      record_application_tags = true
      record_client_address   = true
    }

    maintenance_window {
      day          = var.maintenance_window_day
      hour         = var.maintenance_window_hour
      update_track = var.maintenance_window_update_track
    }

    ip_configuration {
      ipv4_enabled                                  = var.ipv4_enabled
      private_network                               = var.private_network
      enable_private_path_for_google_cloud_services = true
    }

    dynamic "database_flags" {
      for_each = var.database_flags
      content {
        name  = database_flags.key
        value = database_flags.value
      }
    }
  }
}

resource "google_sql_database" "app" {
  name            = var.database_name
  instance        = google_sql_database_instance.this.name
  deletion_policy = var.database_deletion_policy
}

resource "google_sql_user" "app" {
  name            = var.app_username
  instance        = google_sql_database_instance.this.name
  password        = var.app_password
  deletion_policy = var.user_deletion_policy
}
