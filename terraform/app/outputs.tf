output "service_url" {
  value = google_cloud_run_v2_service.github_stats.uri
}

output "refresh_job_name" {
  value = google_cloud_run_v2_job.snapshot_refresh.name
}

output "refresh_schedule_name" {
  value = google_cloud_scheduler_job.snapshot_refresh.name
}
