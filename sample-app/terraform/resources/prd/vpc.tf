resource "google_compute_network" "vpc" {
  name                    = "${var.env}-${var.role}"
  auto_create_subnetworks = "false"
}

resource "google_compute_subnetwork" "asia-northeast1" {
  name          = "${var.env}-${var.role}-${var.region}"
  ip_cidr_range = "10.155.0.0/20"
  network       = "${google_compute_network.vpc.self_link}"
  region        = "${var.region}"
}

resource "google_compute_global_address" "global_static_ip_web" {
  name = "${var.env}-${var.role}-web"
}

resource "google_compute_global_address" "global_static_ip_api" {
  name = "${var.env}-${var.role}-api"
}

resource "google_compute_global_address" "global_static_ip_admin" {
  name = "${var.env}-${var.role}-admin"
}

resource "google_compute_address" "static_ip_nat" {
  name   = "${var.env}-${var.role}-nat${format("%02d", count.index+1)}"
  region = "${var.region}"
  count  = 1
}
