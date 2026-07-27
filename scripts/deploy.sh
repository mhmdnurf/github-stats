#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

project_id="${PROJECT_ID:-mhmdnurf-github-stats}"
region="${REGION:-asia-southeast2}"
service_name="${SERVICE_NAME:-github-stats}"
artifact_repository="${ARTIFACT_REPOSITORY:-github-stats}"
state_bucket="${TF_STATE_BUCKET:-${project_id}-${service_name}-tfstate}"
secret_name="${GITHUB_TOKEN_SECRET_NAME:-github-stats-github-token}"
runtime_service_account="${RUNTIME_SERVICE_ACCOUNT:-github-stats-runtime@${project_id}.iam.gserviceaccount.com}"
refresh_service_account="${REFRESH_SERVICE_ACCOUNT:-github-stats-refresh@${project_id}.iam.gserviceaccount.com}"
scheduler_service_account="${SCHEDULER_SERVICE_ACCOUNT:-github-stats-scheduler@${project_id}.iam.gserviceaccount.com}"
firestore_collection="${FIRESTORE_COLLECTION:-github_stats_snapshots}"
refresh_schedule="${REFRESH_SCHEDULE:-*/15 * * * *}"
refresh_job_name="${service_name}-refresh"
run_initial_refresh="${RUN_INITIAL_REFRESH:-true}"
deploy_mode="${DEPLOY_MODE:-all}"

if [[ -f "${root_dir}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "${root_dir}/.env"
  set +a
fi

: "${GITHUB_USERNAME:?GITHUB_USERNAME must be set in .env or the environment}"

for command in gcloud terraform; do
  if ! command -v "${command}" >/dev/null 2>&1; then
    echo "${command} is required" >&2
    exit 1
  fi
done

terraform_init() {
  local directory="$1"
  local prefix="$2"

  terraform -chdir="${root_dir}/${directory}" init -reconfigure \
    -backend-config="bucket=${state_bucket}" \
    -backend-config="prefix=${prefix}"
}

if [[ "${deploy_mode}" == "all" ]]; then
  gcloud services enable storage.googleapis.com \
    --project="${project_id}"

  if ! gcloud storage buckets describe "gs://${state_bucket}" \
    --project="${project_id}" >/dev/null 2>&1; then
    gcloud storage buckets create "gs://${state_bucket}" \
      --project="${project_id}" \
      --location="${region}" \
      --uniform-bucket-level-access
  fi

  terraform_init "terraform/bootstrap" "bootstrap"

  terraform -chdir="${root_dir}/terraform/bootstrap" apply \
    -auto-approve \
    -var="project_id=${project_id}" \
    -var="region=${region}" \
    -var="service_name=${service_name}" \
    -var="state_bucket=${state_bucket}" \
    -var="retain_legacy_runtime_secret_access=true"

  if [[ -n "${GITHUB_TOKEN:-}" ]]; then
    printf '%s' "${GITHUB_TOKEN}" | \
      gcloud secrets versions add "${secret_name}" \
        --project="${project_id}" \
        --data-file=-
  fi
fi

image_tag="${IMAGE_TAG:-$(git -C "${root_dir}" rev-parse --short HEAD)}"
image="${region}-docker.pkg.dev/${project_id}/${artifact_repository}/${service_name}:${image_tag}"

gcloud builds submit "${root_dir}" \
  --project="${project_id}" \
  --tag="${image}"

terraform_init "terraform/app" "app"

terraform_variables=(
  "-var=project_id=${project_id}"
  "-var=region=${region}"
  "-var=service_name=${service_name}"
  "-var=image=${image}"
  "-var=github_username=${GITHUB_USERNAME}"
  "-var=runtime_service_account=${runtime_service_account}"
  "-var=refresh_service_account=${refresh_service_account}"
  "-var=scheduler_service_account=${scheduler_service_account}"
  "-var=github_token_secret_id=${secret_name}"
  "-var=firestore_collection=${firestore_collection}"
  "-var=refresh_schedule=${refresh_schedule}"
)

if [[ "${run_initial_refresh}" == "true" ]]; then
  terraform -chdir="${root_dir}/terraform/app" apply \
    -auto-approve \
    -target=google_cloud_run_v2_job.snapshot_refresh \
    "${terraform_variables[@]}"

  gcloud run jobs execute "${refresh_job_name}" \
    --project="${project_id}" \
    --region="${region}" \
    --wait
fi

terraform -chdir="${root_dir}/terraform/app" apply \
  -auto-approve \
  "${terraform_variables[@]}"

if [[ "${deploy_mode}" == "all" ]]; then
  terraform -chdir="${root_dir}/terraform/bootstrap" apply \
    -auto-approve \
    -var="project_id=${project_id}" \
    -var="region=${region}" \
    -var="service_name=${service_name}" \
    -var="state_bucket=${state_bucket}" \
    -var="retain_legacy_runtime_secret_access=false"
fi
