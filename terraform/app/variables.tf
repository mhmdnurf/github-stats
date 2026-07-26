variable "project_id" {
  type        = string
  description = "Google Cloud project ID."
  default     = "mhmdnurf-github-stats"
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

variable "github_token_secret_id" {
  type        = string
  description = "Secret Manager secret ID containing GITHUB_TOKEN."
}
