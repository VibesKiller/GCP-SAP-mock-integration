output "service_account_emails" {
  value = {
    for key, service_account in google_service_account.this : key => service_account.email
  }
}

output "service_account_names" {
  value = {
    for key, service_account in google_service_account.this : key => service_account.name
  }
}

output "service_account_members" {
  value = {
    for key, service_account in google_service_account.this : key => "serviceAccount:${service_account.email}"
  }
}
