locals {
  platform_name        = "sap-integration"
  name_prefix          = "${local.platform_name}-${var.environment}"
  kubernetes_namespace = var.kubernetes_namespace
  artifact_repository  = "sap-integration"
  postgresql_username  = "integration_app"
  postgresql_database  = "integration"
  kafka_cluster_id     = "${local.name_prefix}-kafka"
  network_name         = "${local.name_prefix}-vpc"
  subnet_name          = "${local.name_prefix}-subnet"
  gke_cluster_name     = "${local.name_prefix}-gke"
  postgresql_instance  = "${local.name_prefix}-pgsql"

  labels = merge({
    environment = var.environment
    managed_by  = "terraform"
    platform    = local.platform_name
    repository  = "gcp-sap-mock-integration"
    criticality = "business"
  }, var.additional_labels)

  required_services = toset([
    "artifactregistry.googleapis.com",
    "compute.googleapis.com",
    "container.googleapis.com",
    "iam.googleapis.com",
    "logging.googleapis.com",
    "managedkafka.googleapis.com",
    "monitoring.googleapis.com",
    "secretmanager.googleapis.com",
    "servicenetworking.googleapis.com",
    "sqladmin.googleapis.com",
  ])

  kafka_topic_catalog          = yamldecode(file("${path.module}/../../../platform/kafka/topic-catalog.yaml"))
  kafka_consumer_group_catalog = yamldecode(file("${path.module}/../../../platform/kafka/consumer-groups.yaml"))

  workload_ksa_names = {
    sap-mock-api        = "sap-mock-api"
    ingestion-api       = "ingestion-api"
    event-processor     = "event-processor"
    query-api           = "query-api"
    notification-worker = "notification-worker"
  }

  service_accounts = {
    gke-nodes = {
      account_id   = "gke-nodes-${var.environment}"
      display_name = "GKE nodes (${var.environment})"
      description  = "Node service account for the ${var.environment} GKE cluster."
      project_roles = [
        "roles/artifactregistry.reader",
        "roles/logging.logWriter",
        "roles/monitoring.metricWriter",
        "roles/stackdriver.resourceMetadata.writer",
      ]
      kubernetes_service_accounts = []
    }
    sap-mock-api = {
      account_id    = "sap-mock-api-${var.environment}"
      display_name  = "sap-mock-api (${var.environment})"
      description   = "Google service account for the SAP mock API workload."
      project_roles = []
      kubernetes_service_accounts = [{
        namespace            = local.kubernetes_namespace
        service_account_name = local.workload_ksa_names.sap-mock-api
      }]
    }
    ingestion-api = {
      account_id   = "ingestion-api-${var.environment}"
      display_name = "ingestion-api (${var.environment})"
      description  = "Google service account for the ingestion API workload."
      project_roles = [
        "roles/managedkafka.client",
      ]
      kubernetes_service_accounts = [{
        namespace            = local.kubernetes_namespace
        service_account_name = local.workload_ksa_names.ingestion-api
      }]
    }
    event-processor = {
      account_id   = "event-processor-${var.environment}"
      display_name = "event-processor (${var.environment})"
      description  = "Google service account for the Kafka event processor workload."
      project_roles = [
        "roles/managedkafka.client",
      ]
      kubernetes_service_accounts = [{
        namespace            = local.kubernetes_namespace
        service_account_name = local.workload_ksa_names.event-processor
      }]
    }
    query-api = {
      account_id    = "query-api-${var.environment}"
      display_name  = "query-api (${var.environment})"
      description   = "Google service account for the read-only query API workload."
      project_roles = []
      kubernetes_service_accounts = [{
        namespace            = local.kubernetes_namespace
        service_account_name = local.workload_ksa_names.query-api
      }]
    }
    notification-worker = {
      account_id   = "notification-worker-${var.environment}"
      display_name = "notification-worker (${var.environment})"
      description  = "Reserved Google service account for the optional notification worker."
      project_roles = [
        "roles/managedkafka.client",
      ]
      kubernetes_service_accounts = [{
        namespace            = local.kubernetes_namespace
        service_account_name = local.workload_ksa_names.notification-worker
      }]
    }
  }

  kafka_topics = {
    for topic in local.kafka_topic_catalog.topics : topic.name => {
      partition_count    = topic.partitions
      replication_factor = topic.replicationFactor
      configs            = topic.name == "sap.integration.dlq.v1" ? var.kafka_dlq_topic_configs : var.kafka_default_topic_configs
    }
  }

  consumer_group_owners = {
    for consumer_group in local.kafka_consumer_group_catalog.consumerGroups : consumer_group.name => consumer_group.owner
  }

  kafka_client_accounts = distinct(concat(
    [for topic in local.kafka_topic_catalog.topics : topic.owner],
    [for consumer_group in local.kafka_consumer_group_catalog.consumerGroups : consumer_group.owner],
  ))

  workload_identity_bindings_flat = flatten([
    for account_key, account in local.service_accounts : [
      for ksa in try(account.kubernetes_service_accounts, []) : {
        key                  = "${account_key}/${ksa.namespace}/${ksa.service_account_name}"
        account_key          = account_key
        namespace            = ksa.namespace
        service_account_name = ksa.service_account_name
      }
    ]
  ])

  workload_identity_bindings = {
    for binding in local.workload_identity_bindings_flat : binding.key => binding
  }

  kafka_acls = merge(
    {
      for topic in local.kafka_topic_catalog.topics : "topic-${replace(topic.name, ".", "-")}" => {
        acl_id = "topic/${topic.name}"
        acl_entries = concat(
          [
            {
              principal       = "User:${module.iam_service_accounts.service_account_emails[topic.owner]}"
              operation       = "WRITE"
              permission_type = "ALLOW"
              host            = "*"
            },
            {
              principal       = "User:${module.iam_service_accounts.service_account_emails[topic.owner]}"
              operation       = "DESCRIBE"
              permission_type = "ALLOW"
              host            = "*"
            },
          ],
          flatten([
            for consumer_group_name in topic.consumerGroups : [
              {
                principal       = "User:${module.iam_service_accounts.service_account_emails[local.consumer_group_owners[consumer_group_name]]}"
                operation       = "READ"
                permission_type = "ALLOW"
                host            = "*"
              },
              {
                principal       = "User:${module.iam_service_accounts.service_account_emails[local.consumer_group_owners[consumer_group_name]]}"
                operation       = "DESCRIBE"
                permission_type = "ALLOW"
                host            = "*"
              },
            ]
          ])
        )
      }
    },
    {
      for consumer_group in local.kafka_consumer_group_catalog.consumerGroups : "consumer-group-${replace(consumer_group.name, ".", "-")}" => {
        acl_id = "consumerGroup/${consumer_group.name}"
        acl_entries = [
          {
            principal       = "User:${module.iam_service_accounts.service_account_emails[consumer_group.owner]}"
            operation       = "READ"
            permission_type = "ALLOW"
            host            = "*"
          },
          {
            principal       = "User:${module.iam_service_accounts.service_account_emails[consumer_group.owner]}"
            operation       = "DESCRIBE"
            permission_type = "ALLOW"
            host            = "*"
          },
        ]
      }
    },
    {
      cluster = {
        acl_id = "cluster"
        acl_entries = [
          for account_name in local.kafka_client_accounts : {
            principal       = "User:${module.iam_service_accounts.service_account_emails[account_name]}"
            operation       = "DESCRIBE"
            permission_type = "ALLOW"
            host            = "*"
          }
        ]
      }
    }
  )
}

resource "google_project_service" "required" {
  for_each                   = local.required_services
  project                    = var.project_id
  service                    = each.value
  disable_dependent_services = false
  disable_on_destroy         = false
}

resource "random_password" "postgresql_app" {
  length           = 32
  special          = true
  override_special = "_%@"
}

resource "random_password" "sap_api_token" {
  length  = 40
  special = false
}

module "network" {
  source                                 = "../../modules/network"
  network_name                           = local.network_name
  subnet_name                            = local.subnet_name
  region                                 = var.region
  subnet_cidr                            = var.subnet_cidr
  pods_range_name                        = "gke-pods"
  pods_cidr                              = var.gke_pods_cidr
  services_range_name                    = "gke-services"
  services_cidr                          = var.gke_services_cidr
  private_service_range_address          = var.private_service_range_address
  private_service_range_prefix_length    = var.private_service_range_prefix_length
  private_service_access_deletion_policy = var.private_service_access_deletion_policy

  depends_on = [google_project_service.required]
}

module "artifact_registry" {
  source        = "../../modules/artifact-registry"
  region        = var.region
  repository_id = local.artifact_repository

  depends_on = [google_project_service.required]
}

module "iam_service_accounts" {
  source                            = "../../modules/iam-service-accounts"
  project_id                        = var.project_id
  service_accounts                  = local.service_accounts
  create_workload_identity_bindings = false

  depends_on = [google_project_service.required]
}

module "secret_manager" {
  source = "../../modules/secret-manager"
  secrets = {
    postgresql_app_password = {
      secret_id              = "${local.name_prefix}-postgres-app-password"
      labels                 = local.labels
      create_initial_version = true
      secret_data            = random_password.postgresql_app.result
    }
    sap_api_token = {
      secret_id              = "${local.name_prefix}-sap-api-token"
      labels                 = local.labels
      create_initial_version = true
      secret_data            = random_password.sap_api_token.result
    }
  }
  accessor_bindings = {
    postgresql_app_password_event_processor = {
      secret_key = "postgresql_app_password"
      member     = module.iam_service_accounts.service_account_members["event-processor"]
    }
    postgresql_app_password_query_api = {
      secret_key = "postgresql_app_password"
      member     = module.iam_service_accounts.service_account_members["query-api"]
    }
    sap_api_token_ingestion_api = {
      secret_key = "sap_api_token"
      member     = module.iam_service_accounts.service_account_members["ingestion-api"]
    }
  }

  depends_on = [google_project_service.required, module.iam_service_accounts]
}

module "postgresql" {
  source                   = "../../modules/postgresql"
  instance_name            = local.postgresql_instance
  region                   = var.region
  database_name            = local.postgresql_database
  database_deletion_policy = var.postgresql_database_deletion_policy
  app_username             = local.postgresql_username
  user_deletion_policy     = var.postgresql_user_deletion_policy
  app_password             = random_password.postgresql_app.result
  tier                     = var.postgresql_tier
  availability_type        = var.postgresql_availability_type
  disk_size_gb             = var.postgresql_disk_size_gb
  private_network          = module.network.network_self_link
  ipv4_enabled             = false
  deletion_protection      = var.deletion_protection
  labels                   = local.labels

  depends_on = [google_project_service.required, module.network]
}

module "managed_kafka" {
  source     = "../../modules/managed-kafka"
  project_id = var.project_id
  cluster_id = local.kafka_cluster_id
  location   = var.region
  subnet     = module.network.subnet_id
  vcpu_count = var.kafka_vcpu_count
  memory_gib = var.kafka_memory_gib
  labels     = local.labels
  topics     = local.kafka_topics
  acls       = local.kafka_acls

  depends_on = [google_project_service.required, module.network, module.iam_service_accounts]
}

module "gke_cluster" {
  source                     = "../../modules/gke-cluster"
  project_id                 = var.project_id
  cluster_name               = local.gke_cluster_name
  region                     = var.region
  network_self_link          = module.network.network_self_link
  subnet_self_link           = module.network.subnet_self_link
  pods_range_name            = module.network.pods_range_name
  services_range_name        = module.network.services_range_name
  release_channel            = var.gke_release_channel
  master_ipv4_cidr_block     = var.gke_master_ipv4_cidr_block
  master_authorized_networks = var.master_authorized_networks
  machine_type               = var.gke_machine_type
  node_locations             = var.gke_node_locations
  disk_type                  = var.gke_disk_type
  disk_size_gb               = var.gke_disk_size_gb
  min_node_count             = var.gke_min_node_count
  max_node_count             = var.gke_max_node_count
  node_service_account       = module.iam_service_accounts.service_account_emails["gke-nodes"]
  deletion_protection        = var.deletion_protection
  labels                     = local.labels
  network_tags               = ["${local.name_prefix}-gke"]

  depends_on = [google_project_service.required, module.network, module.iam_service_accounts]
}

resource "google_service_account_iam_member" "workload_identity_user" {
  for_each           = local.workload_identity_bindings
  service_account_id = module.iam_service_accounts.service_account_names[each.value.account_key]
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${module.gke_cluster.workload_pool}[${each.value.namespace}/${each.value.service_account_name}]"

  depends_on = [module.gke_cluster]
}
