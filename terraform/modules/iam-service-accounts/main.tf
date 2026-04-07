locals {
  project_role_bindings_flat = flatten([
    for account_key, account in var.service_accounts : [
      for role in try(account.project_roles, []) : {
        key         = "${account_key}/${role}"
        account_key = account_key
        role        = role
      }
    ]
  ])

  project_role_bindings = {
    for binding in local.project_role_bindings_flat : binding.key => binding
  }

  workload_identity_bindings_flat = flatten([
    for account_key, account in var.service_accounts : [
      for ksa in try(account.kubernetes_service_accounts, []) : {
        key                  = "${account_key}/${ksa.namespace}/${ksa.service_account_name}"
        account_key          = account_key
        namespace            = ksa.namespace
        service_account_name = ksa.service_account_name
      }
    ]
  ])

  workload_identity_bindings = {
    for binding in local.workload_identity_bindings_flat : binding.key => binding
  }

  enabled_workload_identity_bindings = var.create_workload_identity_bindings ? local.workload_identity_bindings : {}
}

resource "google_service_account" "this" {
  for_each     = var.service_accounts
  account_id   = each.value.account_id
  display_name = each.value.display_name
  description  = each.value.description
}

resource "google_project_iam_member" "project_role" {
  for_each = local.project_role_bindings
  project  = var.project_id
  role     = each.value.role
  member   = "serviceAccount:${google_service_account.this[each.value.account_key].email}"
}

resource "google_service_account_iam_member" "workload_identity_user" {
  for_each           = local.enabled_workload_identity_bindings
  service_account_id = google_service_account.this[each.value.account_key].name
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[${each.value.namespace}/${each.value.service_account_name}]"
}
