terraform {
  required_version = ">= 1.7.0"

  backend "gcs" {}

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.49"
    }

    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }

    http = {
      source  = "hashicorp/http"
      version = "~> 3.4"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}
