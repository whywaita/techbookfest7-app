resource "google_service_account" "object_viewer" {
  account_id   = "${var.env}-${var.role}-gcs"
  display_name = "${var.env}-${var.role}-gcs"
}
