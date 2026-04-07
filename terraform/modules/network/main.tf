resource "google_compute_network" "this" {
  name                    = var.network_name
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"
}

resource "google_compute_subnetwork" "primary" {
  name                     = var.subnet_name
  ip_cidr_range            = var.subnet_cidr
  region                   = var.region
  network                  = google_compute_network.this.id
  private_ip_google_access = true

  secondary_ip_range {
    range_name    = var.pods_range_name
    ip_cidr_range = var.pods_cidr
  }

  secondary_ip_range {
    range_name    = var.services_range_name
    ip_cidr_range = var.services_cidr
  }

  dynamic "log_config" {
    for_each = var.enable_flow_logs ? [1] : []
    content {
      aggregation_interval = "INTERVAL_10_MIN"
      flow_sampling        = var.flow_logs_sampling
      metadata             = "INCLUDE_ALL_METADATA"
    }
  }
}

resource "google_compute_router" "this" {
  count   = var.create_cloud_nat ? 1 : 0
  name    = coalesce(var.cloud_router_name, "${var.network_name}-router")
  region  = var.region
  network = google_compute_network.this.id
}

resource "google_compute_router_nat" "this" {
  count                              = var.create_cloud_nat ? 1 : 0
  name                               = coalesce(var.cloud_nat_name, "${var.network_name}-nat")
  router                             = google_compute_router.this[0].name
  region                             = var.region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "LIST_OF_SUBNETWORKS"

  subnetwork {
    name                    = google_compute_subnetwork.primary.id
    source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
  }
}

resource "google_compute_global_address" "private_service_range" {
  count         = var.create_private_service_access ? 1 : 0
  name          = coalesce(var.private_service_range_name, "${var.network_name}-private-services")
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  address       = var.private_service_range_address
  prefix_length = var.private_service_range_prefix_length
  network       = google_compute_network.this.id
}

resource "google_service_networking_connection" "private_service_access" {
  count                   = var.create_private_service_access ? 1 : 0
  network                 = google_compute_network.this.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_service_range[0].name]
  deletion_policy         = var.private_service_access_deletion_policy
}
