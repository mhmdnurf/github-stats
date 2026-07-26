# Terraform deployment

`scripts/deploy.sh` is the one-command deployment entry point. It creates the
Terraform state bucket, provisions the bootstrap resources, uploads the GitHub
token from the local `.env` file to Secret Manager, builds the container with
Cloud Build, and deploys the resulting image to Cloud Run.

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
| Artifact Registry repository | `github-stats` |

The script creates a GCS bucket for remote Terraform state before Terraform is
initialized. The bucket is deliberately bootstrapped by the script because a
Terraform backend must already exist before Terraform can store state in it.

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

Cloud Run runs in `asia-southeast2`, listens on port `8080`, scales from zero,
and is capped at one instance. The GitHub token is read by the runtime service
account from Secret Manager.
