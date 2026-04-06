output "db_password_secret_id" {
  value = google_secret_manager_secret.db_password.secret_id
}

output "sap_api_token_secret_id" {
  value = google_secret_manager_secret.sap_api_token.secret_id
}
