resource "google_container_cluster" "this" {
  name     = var.cluster_name
  location = var.region

  network    = var.network_self_link
  subnetwork = var.subnet_self_link

  deletion_protection = var.deletion_protection
  remove_default_node_pool = true
  initial_node_count       = 1

  release_channel {
    channel = var.release_channel
  }

  ip_allocation_policy {
    cluster_secondary_range_name  = var.pods_range_name
    services_secondary_range_name = var.services_range_name
  }

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
}

resource "google_container_node_pool" "primary" {
  name       = "${var.cluster_name}-pool"
  cluster    = google_container_cluster.this.name
  location   = var.region
  node_count = var.node_count

  node_config {
    machine_type    = var.machine_type
    service_account = var.node_service_account
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
    labels = var.labels
  }
}
