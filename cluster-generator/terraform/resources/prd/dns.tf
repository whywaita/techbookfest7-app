resource "google_dns_managed_zone" "prd" {
  name     = "your-domain"
  dns_name = "your-domain.com."
}

resource "google_dns_record_set" "a" {
  name = "${google_dns_managed_zone.prd.dns_name}"
  managed_zone = "${google_dns_managed_zone.prd.name}"
  type = "A"
  ttl  = 300

  rrdatas = ["${google_compute_global_address.global_static_ip_api.address}"]
}
