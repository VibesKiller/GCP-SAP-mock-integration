output "repository_id" {
  value = google_artifact_registry_repository.this.repository_id
}

output "repository_name" {
  value = google_artifact_registry_repository.this.name
}

output "repository_url" {
  value = "${google_artifact_registry_repository.this.location}-docker.pkg.dev/${google_artifact_registry_repository.this.project}/${google_artifact_registry_repository.this.repository_id}"
}
