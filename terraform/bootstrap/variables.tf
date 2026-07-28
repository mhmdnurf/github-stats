variable "project_id" {
  type        = string
  description = "Google Cloud project ID."
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

variable "github_repository" {
  type        = string
  description = "GitHub repository allowed to deploy through Workload Identity Federation."

  validation {
    condition     = can(regex("^[^/]+/[^/]+$", var.github_repository))
    error_message = "github_repository must use the owner/repository format."
  }
}

variable "retain_legacy_runtime_secret_access" {
  type        = bool
  description = "Keep runtime token access until the snapshot-based server rollout succeeds."
  default     = true
}
