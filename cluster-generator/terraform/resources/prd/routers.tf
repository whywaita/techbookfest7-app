resource "google_compute_router" "nat_router" {
  name    = "${var.env}-${var.role}-nat-router"
  network = "${google_compute_network.vpc.name}"
}

resource "google_compute_router_nat" "nat" {
  name                               = "${var.env}-${var.role}-nat-config"
  router                             = "${google_compute_router.nat_router.name}"
  region                             = "${var.region}"
  nat_ip_allocate_option             = "MANUAL_ONLY"
  nat_ips                            = "${google_compute_address.static_ip_nat.*.self_link}"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
  min_ports_per_vm                   = 512
}
