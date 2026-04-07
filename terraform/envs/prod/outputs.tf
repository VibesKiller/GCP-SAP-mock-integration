output "network_name" {
  value = module.network.network_name
}

output "gke_cluster_name" {
  value = module.gke_cluster.cluster_name
}

output "gke_get_credentials_command" {
  value = "gcloud container clusters get-credentials ${module.gke_cluster.cluster_name} --region ${var.region} --project ${var.project_id}"
}

output "gke_workload_pool" {
  value = module.gke_cluster.workload_pool
}

output "artifact_registry_repository_url" {
  value = module.artifact_registry.repository_url
}

output "cloudsql_connection_name" {
  value = module.postgresql.connection_name
}

output "cloudsql_private_ip_address" {
  value = module.postgresql.private_ip_address
}

output "cloudsql_database_name" {
  value = module.postgresql.database_name
}

output "cloudsql_app_username" {
  value = module.postgresql.app_username
}

output "secret_ids" {
  value = module.secret_manager.secret_ids
}

output "service_account_emails" {
  value = module.iam_service_accounts.service_account_emails
}

output "managed_kafka_cluster_name" {
  value = module.managed_kafka.cluster_name
}

output "managed_kafka_bootstrap_address_lookup_command" {
  value = module.managed_kafka.bootstrap_address_lookup_command
}

output "managed_kafka_topic_ids" {
  value = module.managed_kafka.topic_ids
}

output "helm_workload_identity_values" {
  value = {
    for service_name, ksa_name in local.workload_ksa_names : service_name => {
      kubernetes_namespace       = local.kubernetes_namespace
      kubernetes_service_account = ksa_name
      google_service_account     = module.iam_service_accounts.service_account_emails[service_name]
      annotation_key             = "iam.gke.io/gcp-service-account"
      annotation_value           = module.iam_service_accounts.service_account_emails[service_name]
    }
  }
}
