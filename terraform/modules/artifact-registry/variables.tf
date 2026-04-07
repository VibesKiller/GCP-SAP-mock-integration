variable "region" {
  type = string
}

variable "repository_id" {
  type = string
}

variable "format" {
  type    = string
  default = "DOCKER"
}

variable "description" {
  type    = string
  default = "Artifact Registry repository for the SAP integration platform"
}
