output "network_name" {
  value = google_compute_network.this.name
}

output "network_self_link" {
  value = google_compute_network.this.self_link
}

output "subnet_name" {
  value = google_compute_subnetwork.primary.name
}

output "subnet_id" {
  value = google_compute_subnetwork.primary.id
}

output "subnet_self_link" {
  value = google_compute_subnetwork.primary.self_link
}

output "pods_range_name" {
  value = var.pods_range_name
}

output "services_range_name" {
  value = var.services_range_name
}

output "cloud_nat_name" {
  value = try(google_compute_router_nat.this[0].name, null)
}

output "private_service_connection" {
  value = try(google_service_networking_connection.private_service_access[0].peering, null)
}
