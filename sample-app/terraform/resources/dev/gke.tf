provider "google-beta" {
  project = "sample-project-name"
  region  = "asia-northeast1"
}

data "google_compute_network" "vpc_gke" {
  name = "${var.env}-${var.role}"
}

data "google_compute_subnetwork" "asia_northeast1" {
  name   = "${var.env}-${var.role}-${var.region}"
  region = "${var.region}"
}

resource "google_container_cluster" "dev_generator_sample_app" {
  provider = "google-beta"
  name     = "${var.env}-${var.role}"
  location = "${var.location}"

  network    = "${data.google_compute_network.vpc_gke.self_link}"
  subnetwork = "${data.google_compute_subnetwork.asia_northeast1.self_link}"

  ip_allocation_policy {
    use_ip_aliases = true
  }

  cluster_autoscaling {
    enabled = false
  }
  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "0.0.0.0/0"
      display_name = "Public"
    }
  }


  resource_labels = {
    env = "${var.env}"
  }

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true

  initial_node_count = 1

  # Setting an empty username and password explicitly disables basic auth
  master_auth {
    username = ""
    password = ""
  }
}

resource "google_container_node_pool" "dev_generator_sample_app_standard" {
  project    = "sample-project-name"
  name       = "${var.env}-${var.role}-standard"
  location   = "${var.location}"
  cluster    = "${google_container_cluster.dev_generator_sample_app.name}"

  node_count = 1

  node_config {
    preemptible  = false
    machine_type = "n1-standard-1"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes = [
      # gke-default
      "https://www.googleapis.com/auth/devstorage.read_only",

      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/trace.append",
    ]
  }
}

# The following outputs allow authentication and connectivity to the GKE Cluster
# by using certificate-based authentication.
output "client_certificate" {
  value = "${google_container_cluster.dev_generator_sample_app.master_auth.0.client_certificate}"
}

output "client_key" {
  value = "${google_container_cluster.dev_generator_sample_app.master_auth.0.client_key}"
}

output "cluster_ca_certificate" {
  value = "${google_container_cluster.dev_generator_sample_app.master_auth.0.cluster_ca_certificate}"
}
