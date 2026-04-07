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

variable "kubernetes_namespace" {
  type    = string
  default = "sap-integration-prod"
}

variable "subnet_cidr" {
  type    = string
  default = "10.110.0.0/20"
}

variable "gke_pods_cidr" {
  type    = string
  default = "10.120.0.0/16"
}

variable "gke_services_cidr" {
  type    = string
  default = "10.130.0.0/20"
}

variable "private_service_range_prefix_length" {
  type    = number
  default = 16
}

variable "private_service_range_address" {
  type        = string
  default     = "10.140.0.0"
  description = "Base address for Private Service Access. Must not overlap primary subnet or GKE secondary ranges."
}

variable "private_service_access_deletion_policy" {
  type        = string
  default     = "DELETE"
  description = "Deletion policy for the Service Networking private service access connection."
}

variable "gke_master_ipv4_cidr_block" {
  type    = string
  default = "172.16.1.0/28"
}

variable "gke_release_channel" {
  type    = string
  default = "REGULAR"
}

variable "gke_machine_type" {
  type    = string
  default = "e2-standard-4"
}

variable "gke_node_locations" {
  type        = list(string)
  default     = []
  description = "Optional GKE node-pool zones. Leave empty in prod unless a deliberate zonal placement policy is required."
}

variable "gke_disk_type" {
  type        = string
  default     = "pd-balanced"
  description = "Boot disk type for GKE nodes."
}

variable "gke_disk_size_gb" {
  type        = number
  default     = 100
  description = "Boot disk size for each GKE node in GB."
}

variable "gke_min_node_count" {
  type    = number
  default = 2
}

variable "gke_max_node_count" {
  type    = number
  default = 6
}

variable "postgresql_tier" {
  type    = string
  default = "db-custom-4-15360"
}

variable "postgresql_availability_type" {
  type    = string
  default = "REGIONAL"
}

variable "postgresql_disk_size_gb" {
  type    = number
  default = 100
}

variable "postgresql_database_deletion_policy" {
  type        = string
  default     = "DELETE"
  description = "Deletion policy for the application database child resource."
}

variable "postgresql_user_deletion_policy" {
  type        = string
  default     = "DELETE"
  description = "Deletion policy for the application database user child resource."
}

variable "kafka_vcpu_count" {
  type    = number
  default = 6
}

variable "kafka_memory_gib" {
  type    = number
  default = 24
}

variable "deletion_protection" {
  type    = bool
  default = true
}

variable "master_authorized_networks" {
  type = list(object({
    cidr_block   = string
    display_name = string
  }))
  default = []
}

variable "additional_labels" {
  type    = map(string)
  default = {}
}

variable "kafka_default_topic_configs" {
  type = map(string)
  default = {
    "cleanup.policy"      = "delete"
    "min.insync.replicas" = "2"
    "retention.ms"        = "604800000"
  }
}

variable "kafka_dlq_topic_configs" {
  type = map(string)
  default = {
    "cleanup.policy"      = "delete"
    "min.insync.replicas" = "2"
    "retention.ms"        = "1209600000"
  }
}
