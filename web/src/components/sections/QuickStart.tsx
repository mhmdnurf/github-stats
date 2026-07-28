import { Section } from '../ui/Section'
import { CodeBlock } from '../ui/CodeBlock'
import { SITE } from '../../lib/config'

const STEPS = [
  {
    title: '1. Clone the repository',
    code: `git clone ${SITE.repoUrl}.git\ncd github-stats`,
  },
  {
    title: '2. Configure the application',
    code: `cp .env.example .env`,
    note: 'Then set GITHUB_USERNAME, GITHUB_TOKEN, GOOGLE_CLOUD_PROJECT, FIRESTORE_COLLECTION, and HTTP_ADDRESS in .env.',
  },
  {
    title: '3. Authenticate and seed Firestore',
    code: `gcloud auth application-default login\ngo run ./cmd/refresh`,
  },
  {
    title: '4. Start with Docker Compose',
    code: `docker compose up -d --build`,
  },
  {
    title: '5. Verify the service',
    code: `curl http://localhost:9000/health`,
  },
]

export function QuickStart() {
  return (
    <Section
      id="quick-start"
      eyebrow="Get started"
      title="Running locally in five steps"
      description="The server reads Application Default Credentials for Firestore. The GitHub token is used by the refresh job and by live requests to dynamic username endpoints."
    >
      <div className="step-list">
        {STEPS.map((step) => (
          <div className="step" key={step.title}>
            <h3>{step.title}</h3>
            {step.note ? <p className="step-note">{step.note}</p> : null}
            <CodeBlock code={step.code} />
          </div>
        ))}
      </div>
    </Section>
  )
}
