output "network_self_link" {
  value = google_compute_network.this.self_link
}

output "subnet_name" {
  value = google_compute_subnetwork.primary.name
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
