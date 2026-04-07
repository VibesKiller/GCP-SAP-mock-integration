variable "project_id" {
  type = string
}

variable "cluster_id" {
  type = string
}

variable "location" {
  type = string
}

variable "subnet" {
  type = string
}

variable "vcpu_count" {
  type = number
}

variable "memory_gib" {
  type = number
}

variable "rebalance_mode" {
  type    = string
  default = "AUTO_REBALANCE_ON_SCALE_UP"
}

variable "labels" {
  type    = map(string)
  default = {}
}

variable "topics" {
  type = map(object({
    partition_count    = number
    replication_factor = number
    configs            = optional(map(string), {})
  }))
  default = {}
}

variable "acls" {
  type = map(object({
    acl_id = string
    acl_entries = list(object({
      principal       = string
      operation       = string
      permission_type = optional(string, "ALLOW")
      host            = optional(string, "*")
    }))
  }))
  default = {}
}
