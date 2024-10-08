terraform {
  required_providers {
    nexus = {
      source  = "datadrivers/nexus"
      version = "2.4.0"
    }
  }
}

provider "nexus" {
  insecure = true
  # TODO(andrewpollack): This is a locally running repository, being brought up/torn down
  # frequently during development. Fix password once this is running for longer
  # than minutes at a time.
  password = var.nexus_password
  url      = var.nexus_url
  username = var.nexus_username
}

resource "nexus_repository_docker_hosted" "container-repo" {
  name   = "container-repo"
  online = true

  docker {
    force_basic_auth = false
    v1_enabled       = false
    http_port        = var.nexus_docker_repo_port
  }

  storage {
    blob_store_name                = "default"
    strict_content_type_validation = true
    write_policy                   = "ALLOW"
  }
}
