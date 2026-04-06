output "cluster_name" {
  value = module.gke_cluster.cluster_name
}

output "artifact_registry_url" {
  value = module.artifact_registry.repository_url
}

output "cloudsql_connection_name" {
  value = module.postgresql.connection_name
}
