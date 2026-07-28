import { Section } from '../ui/Section'
import { Card } from '../ui/Card'

const FEATURES = [
  { title: 'Native SVG rendering', desc: 'Cards are rendered server-side as SVG. No browser or headless renderer required.' },
  { title: 'Persistent snapshots', desc: 'A scheduled refresh job writes stats to Firestore every 15 minutes.' },
  { title: 'In-memory L1 cache', desc: 'Snapshots are preloaded into memory with stale fallback on storage errors.' },
  { title: 'Multiple themes', desc: 'Ships with default, light, dracula, tokyonight, and gruvbox themes.' },
  { title: 'Cloud Run ready', desc: 'Deploy as a Cloud Run Service with a Cloud Run Job and Scheduler for refreshes.' },
  { title: 'Snapshot-backed profile cards', desc: 'Configured-account card requests are served entirely from Firestore snapshots.' },
]

export function Features() {
  return (
    <Section
      id="features"
      eyebrow="Why GitHub Stats"
      title="Built for reliability, not just looks"
      description="Configured-account cards use persisted snapshots for reliability, while dynamic username endpoints can render public cards on demand."
    >
      <div className="feature-grid">
        {FEATURES.map((feature) => (
          <Card key={feature.title}>
            <h3>{feature.title}</h3>
            <p>{feature.desc}</p>
          </Card>
        ))}
      </div>
    </Section>
  )
}
