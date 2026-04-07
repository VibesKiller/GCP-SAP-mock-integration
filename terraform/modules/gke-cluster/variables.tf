variable "project_id" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "region" {
  type = string
}

variable "network_self_link" {
  type = string
}

variable "subnet_self_link" {
  type = string
}

variable "pods_range_name" {
  type = string
}

variable "services_range_name" {
  type = string
}

variable "release_channel" {
  type    = string
  default = "REGULAR"
}

variable "enable_private_nodes" {
  type    = bool
  default = true
}

variable "enable_private_endpoint" {
  type    = bool
  default = false
}

variable "master_ipv4_cidr_block" {
  type    = string
  default = "172.16.0.0/28"
}

variable "master_authorized_networks" {
  type = list(object({
    cidr_block   = string
    display_name = string
  }))
  default = []
}

variable "machine_type" {
  type    = string
  default = "e2-standard-4"
}

variable "node_locations" {
  type        = list(string)
  default     = []
  description = "Optional zone list for the node pool. Leave empty to let GKE use the default regional placement."
}

variable "disk_type" {
  type    = string
  default = "pd-balanced"
}

variable "disk_size_gb" {
  type    = number
  default = 100
}

variable "min_node_count" {
  type    = number
  default = 1
}

variable "max_node_count" {
  type    = number
  default = 3
}

variable "node_service_account" {
  type = string
}

variable "network_tags" {
  type    = list(string)
  default = []
}

variable "deletion_protection" {
  type    = bool
  default = true
}

variable "labels" {
  type    = map(string)
  default = {}
}
