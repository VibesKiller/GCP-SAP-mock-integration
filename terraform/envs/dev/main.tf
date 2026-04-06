locals {
  name_prefix = "sap-int-${var.environment}"
  labels = {
    environment = var.environment
    managed_by  = "terraform"
    repository  = "gcp-sap-mock-integration"
  }
}

module "network" {
  source              = "../../modules/network"
  network_name        = "${local.name_prefix}-vpc"
  subnet_name         = "${local.name_prefix}-subnet"
  region              = var.region
  subnet_cidr         = "10.10.0.0/20"
  pods_range_name     = "gke-pods"
  pods_cidr           = "10.20.0.0/16"
  services_range_name = "gke-services"
  services_cidr       = "10.30.0.0/20"
}

module "artifact_registry" {
  source        = "../../modules/artifact-registry"
  region        = var.region
  repository_id = "sap-integration"
}

module "secret_manager" {
  source      = "../../modules/secret-manager"
  name_prefix = local.name_prefix
}

module "postgresql" {
  source              = "../../modules/postgresql"
  instance_name       = "${local.name_prefix}-pgsql"
  region              = var.region
  database_name       = "integration"
  username            = "integration_app"
  password            = var.db_password
  availability_type   = "ZONAL"
  deletion_protection = false
  ipv4_enabled        = true
}

module "gke_cluster" {
  source               = "../../modules/gke-cluster"
  project_id           = var.project_id
  cluster_name         = "${local.name_prefix}-gke"
  region               = var.region
  network_self_link    = module.network.network_self_link
  subnet_self_link     = module.network.subnet_self_link
  pods_range_name      = module.network.pods_range_name
  services_range_name  = module.network.services_range_name
  node_service_account = "default"
  deletion_protection  = false
  labels               = local.labels
}
