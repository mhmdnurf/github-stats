locals {
  refresh_job_name     = "${var.service_name}-refresh"
  refresh_trigger_name = "${var.service_name}-refresh-schedule"
}

resource "google_cloud_run_v2_service" "github_stats" {
  name     = var.service_name
  location = var.region
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = var.runtime_service_account

    scaling {
      min_instance_count = 1
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
        startup_cpu_boost = true
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
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
      }

      env {
        name  = "FIRESTORE_COLLECTION"
        value = var.firestore_collection
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

resource "google_cloud_run_v2_job" "snapshot_refresh" {
  name     = local.refresh_job_name
  location = var.region

  template {
    task_count  = 1
    parallelism = 1

    template {
      service_account = var.refresh_service_account
      max_retries     = 1
      timeout         = "120s"

      containers {
        image   = var.image
        command = ["/usr/local/bin/github-stats-refresh"]

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
          name  = "GOOGLE_CLOUD_PROJECT"
          value = var.project_id
        }

        env {
          name  = "FIRESTORE_COLLECTION"
          value = var.firestore_collection
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
}

resource "google_cloud_run_v2_job_iam_member" "scheduler_invoker" {
  project  = var.project_id
  location = google_cloud_run_v2_job.snapshot_refresh.location
  name     = google_cloud_run_v2_job.snapshot_refresh.name
  role     = "roles/run.invoker"
  member   = "serviceAccount:${var.scheduler_service_account}"
}

resource "google_cloud_scheduler_job" "snapshot_refresh" {
  project          = var.project_id
  region           = var.region
  name             = local.refresh_trigger_name
  description      = "Refresh GitHub stats snapshots in Firestore"
  schedule         = var.refresh_schedule
  time_zone        = "Etc/UTC"
  attempt_deadline = "60s"

  retry_config {
    retry_count          = 3
    min_backoff_duration = "5s"
    max_backoff_duration = "60s"
    max_doublings        = 2
  }

  http_target {
    http_method = "POST"
    uri         = "https://run.googleapis.com/v2/projects/${var.project_id}/locations/${google_cloud_run_v2_job.snapshot_refresh.location}/jobs/${google_cloud_run_v2_job.snapshot_refresh.name}:run"
    body        = base64encode("{}")

    headers = {
      "Content-Type" = "application/json"
    }

    oauth_token {
      service_account_email = var.scheduler_service_account
      scope                 = "https://www.googleapis.com/auth/cloud-platform"
    }
  }

  depends_on = [
    google_cloud_run_v2_job_iam_member.scheduler_invoker,
  ]
}
