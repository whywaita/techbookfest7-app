terraform {
  backend "gcs" {
    bucket = "cluster-generator"
    prefix = "prd"
  }
}
