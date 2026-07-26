output "artifact_repository" {
  value = google_artifact_registry_repository.images.repository_id
}

output "runtime_service_account" {
  value = google_service_account.runtime.email
}

output "deployer_service_account" {
  value = google_service_account.deployer.email
}

output "github_token_secret_id" {
  value = google_secret_manager_secret.github_token.secret_id
}

output "workload_identity_provider" {
  value = google_iam_workload_identity_pool_provider.github_actions.name
}
