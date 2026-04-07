output "instance_name" {
  value = google_sql_database_instance.this.name
}

output "connection_name" {
  value = google_sql_database_instance.this.connection_name
}

output "database_name" {
  value = google_sql_database.app.name
}

output "private_ip_address" {
  value = google_sql_database_instance.this.private_ip_address
}

output "public_ip_address" {
  value = google_sql_database_instance.this.public_ip_address
}

output "app_username" {
  value = google_sql_user.app.name
}
