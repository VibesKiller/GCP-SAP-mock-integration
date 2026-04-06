resource "google_secret_manager_secret" "db_password" {
  secret_id = "${var.name_prefix}-db-password"

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "sap_api_token" {
  secret_id = "${var.name_prefix}-sap-api-token"

  replication {
    auto {}
  }
}
