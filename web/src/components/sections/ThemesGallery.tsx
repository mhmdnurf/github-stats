import { useState } from 'react'
import { Section } from '../ui/Section'
import { Badge } from '../ui/Badge'
import { Card } from '../ui/Card'
import { CodeBlock } from '../ui/CodeBlock'
import { THEMES, statsCardUrl, languagesCardUrl, type ThemeName } from '../../lib/config'

export function ThemesGallery() {
  const [active, setActive] = useState<ThemeName>('default')

  return (
    <Section
      id="themes"
      eyebrow="Themes"
      title="Five built-in card themes"
      description="Unknown themes return an HTTP 400 Bad Request response. Select a theme to preview it live."
    >
      <div className="theme-picker">
        {THEMES.map((theme) => (
          <button
            key={theme}
            type="button"
            onClick={() => setActive(theme)}
            className="theme-picker-item"
            aria-pressed={active === theme}
          >
            <Badge active={active === theme}>{theme}</Badge>
          </button>
        ))}
      </div>

      <Card className="theme-preview-card">
        <div className="theme-preview-grid">
          <img src={statsCardUrl(active)} alt={`Statistics card, ${active} theme`} loading="lazy" />
          <img src={languagesCardUrl(active)} alt={`Languages card, ${active} theme`} loading="lazy" />
        </div>
        <CodeBlock language="text" code={`https://your-domain.example/stats?theme=${active}`} />
      </Card>
    </Section>
  )
}
