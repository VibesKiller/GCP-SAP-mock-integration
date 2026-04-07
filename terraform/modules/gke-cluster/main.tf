resource "google_container_cluster" "this" {
  name           = var.cluster_name
  location       = var.region
  node_locations = length(var.node_locations) > 0 ? var.node_locations : null

  network    = var.network_self_link
  subnetwork = var.subnet_self_link

  deletion_protection      = var.deletion_protection
  remove_default_node_pool = true
  initial_node_count       = 1

  node_config {
    machine_type    = var.machine_type
    disk_type       = var.disk_type
    disk_size_gb    = var.disk_size_gb
    service_account = var.node_service_account
    image_type      = "COS_CONTAINERD"
    labels          = var.labels
    tags            = var.network_tags

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]

    shielded_instance_config {
      enable_integrity_monitoring = true
      enable_secure_boot          = true
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }

  release_channel {
    channel = var.release_channel
  }

  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"

  ip_allocation_policy {
    cluster_secondary_range_name  = var.pods_range_name
    services_secondary_range_name = var.services_range_name
  }

  private_cluster_config {
    enable_private_nodes    = var.enable_private_nodes
    enable_private_endpoint = var.enable_private_endpoint
    master_ipv4_cidr_block  = var.master_ipv4_cidr_block
  }

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  addons_config {
    horizontal_pod_autoscaling {
      disabled = false
    }

    http_load_balancing {
      disabled = false
    }
  }

  vertical_pod_autoscaling {
    enabled = true
  }

  dynamic "master_authorized_networks_config" {
    for_each = length(var.master_authorized_networks) > 0 ? [1] : []
    content {
      dynamic "cidr_blocks" {
        for_each = var.master_authorized_networks
        content {
          cidr_block   = cidr_blocks.value.cidr_block
          display_name = cidr_blocks.value.display_name
        }
      }
    }
  }

  resource_labels = var.labels
}

resource "google_container_node_pool" "primary" {
  name           = "${var.cluster_name}-primary"
  cluster        = google_container_cluster.this.name
  location       = var.region
  node_locations = length(var.node_locations) > 0 ? var.node_locations : null

  autoscaling {
    min_node_count = var.min_node_count
    max_node_count = var.max_node_count
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }

  node_config {
    machine_type    = var.machine_type
    disk_type       = var.disk_type
    disk_size_gb    = var.disk_size_gb
    service_account = var.node_service_account
    image_type      = "COS_CONTAINERD"
    labels          = var.labels
    tags            = var.network_tags

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
    ]

    shielded_instance_config {
      enable_integrity_monitoring = true
      enable_secure_boot          = true
    }

    workload_metadata_config {
      mode = "GKE_METADATA"
    }
  }
}
