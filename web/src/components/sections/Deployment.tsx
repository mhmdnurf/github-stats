import { Section } from "../ui/Section";
import { Card } from "../ui/Card";
import { CodeBlock } from "../ui/CodeBlock";

export function Deployment() {
	return (
		<Section
			id="deployment"
			eyebrow="Deployment"
			title="Choose the approach that fits your infrastructure"
		>
			<div className="feature-grid">
				<Card>
					<h3>Docker Compose</h3>
					<p>
						Best for local use or an existing server. Requires Docker, Docker
						Compose, and a Firestore database.
					</p>
					<CodeBlock code="docker compose up -d --build" />
				</Card>
				<Card>
					<h3>Google Cloud Run + Terraform</h3>
					<p>
						Managed public hosting with GitHub Actions deployment. Requires a
						billed GCP project, gcloud, and Terraform.
					</p>
					<CodeBlock code="./scripts/deploy.sh" />
				</Card>
			</div>
		</Section>
	);
}
