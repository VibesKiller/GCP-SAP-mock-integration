variable "instance_name" {
  type = string
}

variable "region" {
  type = string
}

variable "database_name" {
  type = string
}

variable "username" {
  type = string
}

variable "password" {
  type      = string
  sensitive = true
}

variable "tier" {
  type    = string
  default = "db-custom-2-7680"
}

variable "availability_type" {
  type    = string
  default = "ZONAL"
}

variable "disk_size_gb" {
  type    = number
  default = 50
}

variable "private_network" {
  type    = string
  default = null
}

variable "ipv4_enabled" {
  type    = bool
  default = false
}

variable "deletion_protection" {
  type    = bool
  default = true
}
