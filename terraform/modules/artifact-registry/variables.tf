variable "region" {
  type = string
}

variable "repository_id" {
  type = string
}

variable "description" {
  type    = string
  default = "Artifact Registry repository for platform services"
}
