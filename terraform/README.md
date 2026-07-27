# Terraform deployment

`scripts/deploy.sh` is the one-command deployment entry point. It creates the
Terraform state bucket, provisions the bootstrap resources, uploads the GitHub
token from the local `.env` file to Secret Manager, builds the container with
Cloud Build, updates and executes the snapshot refresh job to seed Firestore,
and only then deploys the resulting server image to Cloud Run. During the
first cutover, the previous runtime keeps its token access until the new
revision deploys successfully; the script removes that legacy access last.

## Prerequisites

- Authenticated `gcloud` CLI with permission to create project resources
- Terraform 1.7 or newer
- Billing enabled for `mhmdnurf-github-stats`
- A local `.env` containing `GITHUB_USERNAME` and `GITHUB_TOKEN`

Before the first deployment, authenticate both the Google Cloud CLI and the
Terraform Google provider:

```shell
gcloud auth login
gcloud auth application-default login
gcloud config set project mhmdnurf-github-stats
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
| Project | `mhmdnurf-github-stats` |
| Region | `asia-southeast2` |
| Cloud Run service | `github-stats` |
| Cloud Run refresh job | `github-stats-refresh` |
| Snapshot refresh schedule | Every 15 minutes |
| Artifact Registry repository | `github-stats` |

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
| `GITHUB_USERNAME` | GitHub account shown by the cards |
| `GCP_WIF_PROVIDER` | `workload_identity_provider` output |
| `GCP_DEPLOY_SERVICE_ACCOUNT` | `deployer_service_account` output |

The workflow in `.github/workflows/deploy.yml` then deploys on pushes to
`main`. It uses GitHub OIDC and Workload Identity Federation; it does not store
a Google service-account key or the GitHub API token in GitHub Actions.

## Deployment settings

Cloud Run runs in `asia-southeast2`, listens on port `8080`, keeps one instance
warm, and is capped at one instance. The server identity has read-only
Firestore access and does not receive the GitHub token. The refresh job runs
every 15 minutes with a separate writer identity that can read the token from
Secret Manager.

Set `RUN_INITIAL_REFRESH=false` to skip the pre-deployment refresh only when
all four snapshots already exist in Firestore. Otherwise the new server
revision intentionally fails its startup preload. Set `REFRESH_SCHEDULE` to
override the default `*/15 * * * *` cron schedule.
