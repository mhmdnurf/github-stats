variable "project_id" {
  type        = string
  description = "Google Cloud project ID."
  default     = "mhmdnurf-github-stats"
}

variable "region" {
  type        = string
  description = "Region for Artifact Registry and Cloud Run."
  default     = "asia-southeast2"
}

variable "service_name" {
  type        = string
  description = "Cloud Run service name."
  default     = "github-stats"
}

variable "state_bucket" {
  type        = string
  description = "Existing GCS bucket used for Terraform state."
  default     = "mhmdnurf-github-stats-github-stats-tfstate"
}

variable "github_repository" {
  type        = string
  description = "GitHub repository allowed to deploy through Workload Identity Federation."
  default     = "mhmdnurf/github-stats"
}

variable "retain_legacy_runtime_secret_access" {
  type        = bool
  description = "Keep runtime token access until the snapshot-based server rollout succeeds."
  default     = true
}
