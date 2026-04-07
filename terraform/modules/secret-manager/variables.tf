variable "secrets" {
  type = map(object({
    secret_id              = string
    labels                 = optional(map(string), {})
    create_initial_version = optional(bool, false)
    secret_data            = optional(string)
  }))
}

variable "accessor_bindings" {
  type = map(object({
    secret_key = string
    member     = string
  }))
  default     = {}
  description = "Secret-level IAM accessor bindings keyed by static Terraform addresses."
}
