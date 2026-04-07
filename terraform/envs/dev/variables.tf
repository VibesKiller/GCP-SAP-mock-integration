variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "europe-west6"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "kubernetes_namespace" {
  type    = string
  default = "sap-integration-dev"
}

variable "subnet_cidr" {
  type    = string
  default = "10.10.0.0/20"
}

variable "gke_pods_cidr" {
  type    = string
  default = "10.20.0.0/16"
}

variable "gke_services_cidr" {
  type    = string
  default = "10.50.0.0/20"
}

variable "private_service_range_prefix_length" {
  type    = number
  default = 16
}

variable "private_service_range_address" {
  type        = string
  default     = "10.30.0.0"
  description = "Base address for Private Service Access. Must not overlap primary subnet or GKE secondary ranges."
}

variable "private_service_access_deletion_policy" {
  type        = string
  default     = "ABANDON"
  description = "Dev teardown helper: abandon the Service Networking connection if GCP producer services release it asynchronously."
}

variable "gke_master_ipv4_cidr_block" {
  type    = string
  default = "172.16.0.16/28"
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
  description = "Optional GKE node-pool zones. For dev, setting one zone keeps the worker footprint within small quota limits."
}

variable "gke_disk_type" {
  type        = string
  default     = "pd-standard"
  description = "Boot disk type for GKE nodes. Dev defaults to pd-standard to avoid small regional SSD quota limits."
}

variable "gke_disk_size_gb" {
  type        = number
  default     = 50
  description = "Boot disk size for each GKE node in GB."
}

variable "gke_min_node_count" {
  type    = number
  default = 1
}

variable "gke_max_node_count" {
  type    = number
  default = 3
}

variable "postgresql_tier" {
  type    = string
  default = "db-custom-2-7680"
}

variable "postgresql_availability_type" {
  type    = string
  default = "ZONAL"
}

variable "postgresql_disk_size_gb" {
  type    = number
  default = 50
}

variable "postgresql_database_deletion_policy" {
  type        = string
  default     = "ABANDON"
  description = "Dev destroy helper: abandon the database child resource and let Cloud SQL instance deletion remove it."
}

variable "postgresql_user_deletion_policy" {
  type        = string
  default     = "ABANDON"
  description = "Dev destroy helper: abandon the user child resource and let Cloud SQL instance deletion remove it."
}

variable "kafka_vcpu_count" {
  type    = number
  default = 3
}

variable "kafka_memory_gib" {
  type    = number
  default = 12
}

variable "deletion_protection" {
  type    = bool
  default = false
}

variable "master_authorized_networks" {
  type = list(object({
    cidr_block   = string
    display_name = string
  }))
  default = []
}

variable "auto_detect_workstation_public_ip" {
  type        = bool
  default     = true
  description = "When true, Terraform detects the current workstation public IPv4 address and authorizes it for the GKE control plane."
}

variable "workstation_public_ip_lookup_url" {
  type        = string
  default     = "https://checkip.amazonaws.com"
  description = "HTTP endpoint returning the caller public IPv4 address as plain text."
}

variable "auto_detected_workstation_display_name" {
  type        = string
  default     = "terraform-workstation"
  description = "Display name used for the auto-detected workstation CIDR in GKE master authorized networks."
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
