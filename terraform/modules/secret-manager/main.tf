locals {
  versioned_secrets = {
    for key, secret in var.secrets : key => secret
    if try(secret.create_initial_version, false)
  }
}

resource "google_secret_manager_secret" "this" {
  for_each  = var.secrets
  secret_id = each.value.secret_id
  labels    = try(each.value.labels, {})

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "this" {
  for_each    = local.versioned_secrets
  secret      = google_secret_manager_secret.this[each.key].id
  secret_data = each.value.secret_data
}

resource "google_secret_manager_secret_iam_member" "accessor" {
  for_each  = var.accessor_bindings
  secret_id = google_secret_manager_secret.this[each.value.secret_key].secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = each.value.member
}
