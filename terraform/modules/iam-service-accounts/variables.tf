variable "project_id" {
  type = string
}

variable "service_accounts" {
  type = map(object({
    account_id    = string
    display_name  = string
    description   = string
    project_roles = optional(list(string), [])
    kubernetes_service_accounts = optional(list(object({
      namespace            = string
      service_account_name = string
    })), [])
  }))
}

variable "create_workload_identity_bindings" {
  type        = bool
  default     = true
  description = "When true, create IAM bindings between Kubernetes service accounts and Google service accounts. Set to false when the GKE workload pool is created later in the same stack."
}
