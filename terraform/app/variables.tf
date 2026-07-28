variable "project_id" {
  type        = string
  description = "Google Cloud project ID."
}

variable "region" {
  type        = string
  description = "Cloud Run region."
  default     = "asia-southeast2"
}

variable "service_name" {
  type        = string
  description = "Cloud Run service name."
  default     = "github-stats"
}

variable "image" {
  type        = string
  description = "Immutable container image reference to deploy."
}

variable "github_username" {
  type        = string
  description = "GitHub username displayed by the cards."
}

variable "runtime_service_account" {
  type        = string
  description = "Service account used by the Cloud Run revision."
}

variable "refresh_service_account" {
  type        = string
  description = "Service account used by the snapshot refresh job."
}

variable "scheduler_service_account" {
  type        = string
  description = "Service account used by Cloud Scheduler to invoke the refresh job."
}

variable "github_token_secret_id" {
  type        = string
  description = "Secret Manager secret ID containing GITHUB_TOKEN."
}

variable "firestore_collection" {
  type        = string
  description = "Firestore collection containing persistent snapshots."
  default     = "github_stats_snapshots"

  validation {
    condition     = length(trimspace(var.firestore_collection)) > 0
    error_message = "firestore_collection must not be empty."
  }
}

variable "refresh_schedule" {
  type        = string
  description = "Cron schedule used to refresh persistent snapshots."
  default     = "*/15 * * * *"

  validation {
    condition     = length(trimspace(var.refresh_schedule)) > 0
    error_message = "refresh_schedule must not be empty."
  }
}
