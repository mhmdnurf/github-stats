# Terraform deployment

`scripts/deploy.sh` is the one-command deployment entry point. It creates the
Terraform state bucket, provisions the bootstrap resources, uploads the GitHub
token from the local `.env` file to Secret Manager, builds the container with
Cloud Build using a dedicated least-privilege builder service account, updates
and executes the snapshot refresh job to seed Firestore, and only then deploys
the resulting server image to Cloud Run. Application-only deployments assume
the bootstrap stack has already granted the runtime and refresh service accounts
access to the GitHub token.

## Prerequisites

- Authenticated `gcloud` CLI with permission to create project resources
- Terraform 1.7 or newer
- Billing enabled for the target Google Cloud project
- A local `.env` based on `.env.example`

Load the deployment settings from `.env`:

```shell
set -a
source .env
set +a
```

Before the first deployment, authenticate both the Google Cloud CLI and the
Terraform Google provider:

```shell
gcloud auth login
gcloud auth application-default login
gcloud config set project "$PROJECT_ID"
```

## First deployment

Run from the repository root:

```shell
chmod +x scripts/deploy.sh
./scripts/deploy.sh
```

The defaults are:

| Setting | Value |
|---|---|
| Project | Required through `PROJECT_ID` or `GOOGLE_CLOUD_PROJECT` |
| Region | `asia-southeast2` |
| Cloud Run service | `github-stats` |
| Cloud Run refresh job | `github-stats-refresh` |
| Snapshot refresh schedule | Every 15 minutes |
| Artifact Registry repository | `github-stats` |
| Cloud Build service account | `github-stats-builder` |

The script creates a GCS bucket for remote Terraform state before Terraform is
initialized. The bucket is deliberately bootstrapped by the script because a
Terraform backend must already exist before Terraform can store state in it.

## Verify deployment

Get the public Cloud Run URL:

```shell
gcloud run services describe github-stats \
  --region=asia-southeast2 \
  --format='value(status.url)'
```

Check the health endpoint:

```shell
curl "$(gcloud run services describe github-stats \
  --region=asia-southeast2 \
  --format='value(status.url)')/health"
```

Verify that the initial refresh execution succeeded:

```shell
gcloud run jobs executions list \
  --job=github-stats-refresh \
  --region=asia-southeast2 \
  --limit=1
```

## GitHub Actions setup

After the first local deployment, obtain the two bootstrap outputs:

```shell
terraform -chdir=terraform/bootstrap output -raw workload_identity_provider
terraform -chdir=terraform/bootstrap output -raw deployer_service_account
```

Add these repository variables in GitHub:

| Variable | Value |
|---|---|
| `GCP_PROJECT_ID` | Google Cloud project ID |
| `GH_USERNAME` | GitHub account shown by the cards |
| `GCP_WIF_PROVIDER` | `workload_identity_provider` output |
| `GCP_DEPLOY_SERVICE_ACCOUNT` | `deployer_service_account` output |
| `GCP_REGION` | Optional; defaults to `asia-southeast2` |
| `GCP_SERVICE_NAME` | Optional; defaults to `github-stats` |
| `GCP_TF_STATE_BUCKET` | Optional custom state bucket name |

> [!NOTE]
> GitHub rejects repository variable names starting with `GITHUB_`, so the
> username variable is named `GH_USERNAME` instead.

The workflow in `.github/workflows/deploy.yml` then deploys on pushes to
`main`. It uses GitHub OIDC and Workload Identity Federation; it does not store
a Google service-account key or the GitHub API token in GitHub Actions. Cloud
Build runs as the dedicated `github-stats-builder` service account instead of
the project-wide Compute Engine default service account.

If the deployment predates the GCS backend configuration, back up and migrate
both local state files once before running the workflow. Fresh installations
skip this migration:

```shell
cp terraform/bootstrap/terraform.tfstate \
  terraform/bootstrap/terraform.tfstate.backup-manual
cp terraform/app/terraform.tfstate \
  terraform/app/terraform.tfstate.backup-manual
terraform -chdir=terraform/bootstrap init -migrate-state \
  -backend-config="bucket=${TF_STATE_BUCKET}" \
  -backend-config="prefix=bootstrap"
terraform -chdir=terraform/app init -migrate-state \
  -backend-config="bucket=${TF_STATE_BUCKET}" \
  -backend-config="prefix=app"
```

When bootstrap IAM resources change, apply the bootstrap stack once with an
administrator identity before rerunning the application-only workflow:

```shell
terraform -chdir=terraform/bootstrap init -reconfigure \
  -backend-config="bucket=${TF_STATE_BUCKET}" \
  -backend-config="prefix=bootstrap"
terraform -chdir=terraform/bootstrap plan \
  -var="project_id=${PROJECT_ID}" \
  -var="github_repository=${GITHUB_REPOSITORY}"
terraform -chdir=terraform/bootstrap apply \
  -var="project_id=${PROJECT_ID}" \
  -var="github_repository=${GITHUB_REPOSITORY}"
```

For the complete fork setup, including `.env`, first bootstrap, GitHub Actions,
verification, upgrades, and state migration, see
the [self-hosting and deployment section](../README.md#deployment) in the main
README.

## Deployment settings

Cloud Run runs in `asia-southeast2`, listens on port `8080`, keeps one instance
warm, and is capped at one instance. The server identity has read-only
Firestore access and reads `GITHUB_TOKEN` from Secret Manager, which it needs
to serve the `/{username}/stats` and `/{username}/languages` endpoints (these
fetch live GitHub data on demand rather than reading a Firestore snapshot).
The refresh job runs every 15 minutes with a separate writer identity that
also reads the token from Secret Manager.

Set `RUN_INITIAL_REFRESH=false` to skip the pre-deployment refresh only when
all four snapshots already exist in Firestore. Otherwise the new server
revision intentionally fails its startup preload. Set `REFRESH_SCHEDULE` to
override the default `*/15 * * * *` cron schedule.
