terraform {
  backend "gcs" {
    bucket = "generator-sample-app"
    prefix  = "dev"
  }
}
