output "cluster_id" {
  value = google_managed_kafka_cluster.this.cluster_id
}

output "cluster_name" {
  value = google_managed_kafka_cluster.this.name
}

output "cluster_state" {
  value = google_managed_kafka_cluster.this.state
}

output "topic_ids" {
  value = {
    for key, topic in google_managed_kafka_topic.this : key => topic.id
  }
}

output "acl_ids" {
  value = {
    for key, acl in google_managed_kafka_acl.this : key => acl.id
  }
}

output "bootstrap_address_lookup_command" {
  value = "gcloud managed-kafka clusters describe ${google_managed_kafka_cluster.this.cluster_id} --location=${var.location} --project=${var.project_id} --format=\"value(bootstrapAddress)\""
}
