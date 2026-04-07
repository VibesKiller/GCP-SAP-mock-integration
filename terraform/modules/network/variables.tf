variable "network_name" {
  type = string
}

variable "subnet_name" {
  type = string
}

variable "region" {
  type = string
}

variable "subnet_cidr" {
  type = string
}

variable "pods_range_name" {
  type = string
}

variable "pods_cidr" {
  type = string
}

variable "services_range_name" {
  type = string
}

variable "services_cidr" {
  type = string
}

variable "enable_flow_logs" {
  type    = bool
  default = true
}

variable "flow_logs_sampling" {
  type    = number
  default = 0.5
}

variable "create_cloud_nat" {
  type    = bool
  default = true
}

variable "cloud_router_name" {
  type    = string
  default = null
}

variable "cloud_nat_name" {
  type    = string
  default = null
}

variable "create_private_service_access" {
  type    = bool
  default = true
}

variable "private_service_access_deletion_policy" {
  type        = string
  default     = "DELETE"
  description = "Deletion policy for the private service access connection. Use ABANDON for teardown-prone dev environments when producer services release the connection asynchronously."
}

variable "private_service_range_name" {
  type    = string
  default = null
}

variable "private_service_range_address" {
  type        = string
  default     = null
  description = "Optional base address for the Private Service Access VPC peering range. Set it explicitly to avoid overlaps with subnet and GKE secondary ranges."
}

variable "private_service_range_prefix_length" {
  type    = number
  default = 16
}
