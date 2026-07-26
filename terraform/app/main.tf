resource "google_cloud_run_v2_service" "github_stats" {
  name     = var.service_name
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = var.runtime_service_account

    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }

    containers {
      image = var.image

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }

      env {
        name  = "GITHUB_USERNAME"
        value = var.github_username
      }

      env {
        name  = "HTTP_ADDRESS"
        value = ":8080"
      }

      env {
        name = "GITHUB_TOKEN"

        value_source {
          secret_key_ref {
            secret  = var.github_token_secret_id
            version = "latest"
          }
        }
      }
    }
  }
}

resource "google_cloud_run_v2_service_iam_member" "public_invoker" {
  location = google_cloud_run_v2_service.github_stats.location
  name     = google_cloud_run_v2_service.github_stats.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
