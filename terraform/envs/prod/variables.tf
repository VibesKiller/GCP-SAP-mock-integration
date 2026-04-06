variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "europe-west6"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "node_service_account" {
  type = string
}
