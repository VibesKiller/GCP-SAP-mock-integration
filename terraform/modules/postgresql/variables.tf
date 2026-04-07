variable "instance_name" {
  type = string
}

variable "region" {
  type = string
}

variable "database_version" {
  type    = string
  default = "POSTGRES_16"
}

variable "edition" {
  type    = string
  default = "ENTERPRISE"
}

variable "database_name" {
  type = string
}

variable "database_deletion_policy" {
  type        = string
  default     = "DELETE"
  description = "Deletion policy for the application database resource. Use ABANDON when the parent instance destroy should remove application-owned objects."
}

variable "app_username" {
  type = string
}

variable "user_deletion_policy" {
  type        = string
  default     = "DELETE"
  description = "Deletion policy for the application database user resource. Use ABANDON when application-owned objects prevent dropping the role before instance destroy."
}

variable "app_password" {
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

variable "disk_type" {
  type    = string
  default = "PD_SSD"
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

variable "backup_enabled" {
  type    = bool
  default = true
}

variable "point_in_time_recovery_enabled" {
  type    = bool
  default = true
}

variable "maintenance_window_day" {
  type    = number
  default = 7
}

variable "maintenance_window_hour" {
  type    = number
  default = 3
}

variable "maintenance_window_update_track" {
  type    = string
  default = "stable"
}

variable "database_flags" {
  type = map(string)
  default = {
    log_min_duration_statement = "500"
    max_connections            = "200"
  }
}

variable "deletion_protection" {
  type    = bool
  default = true
}

variable "labels" {
  type    = map(string)
  default = {}
}
