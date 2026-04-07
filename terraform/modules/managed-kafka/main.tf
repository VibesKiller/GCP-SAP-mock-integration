locals {
  memory_bytes = floor(var.memory_gib * 1024 * 1024 * 1024)
}

resource "google_managed_kafka_cluster" "this" {
  project    = var.project_id
  cluster_id = var.cluster_id
  location   = var.location
  labels     = var.labels

  capacity_config {
    vcpu_count   = var.vcpu_count
    memory_bytes = local.memory_bytes
  }

  gcp_config {
    access_config {
      network_configs {
        subnet = var.subnet
      }
    }
  }

  rebalance_config {
    mode = var.rebalance_mode
  }
}

resource "google_managed_kafka_topic" "this" {
  for_each           = var.topics
  project            = var.project_id
  topic_id           = each.key
  cluster            = google_managed_kafka_cluster.this.cluster_id
  location           = var.location
  partition_count    = each.value.partition_count
  replication_factor = each.value.replication_factor
  configs            = try(each.value.configs, {})
}

resource "google_managed_kafka_acl" "this" {
  for_each = var.acls

  project  = var.project_id
  acl_id   = each.value.acl_id
  cluster  = google_managed_kafka_cluster.this.cluster_id
  location = var.location

  dynamic "acl_entries" {
    for_each = each.value.acl_entries
    content {
      principal       = acl_entries.value.principal
      operation       = acl_entries.value.operation
      permission_type = try(acl_entries.value.permission_type, "ALLOW")
      host            = try(acl_entries.value.host, "*")
    }
  }
}
