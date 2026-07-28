import { Container } from '../ui/Container'
import { Icon } from '../ui/Icon'
import { LinkButton } from '../ui/Button'
import { ThemeToggle } from '../ui/ThemeToggle'
import { SITE } from '../../lib/config'

const NAV_LINKS = [
  { href: '#quick-start', label: 'Quick Start' },
  { href: '#usage', label: 'Usage' },
  { href: '#themes', label: 'Themes' },
  { href: '#api', label: 'API' },
  { href: '#deployment', label: 'Deployment' },
]

export function Header() {
  return (
    <header className="site-header">
      <Container>
        <div className="site-header-inner">
          <a href="#top" className="brand">
            <Icon name="documentation-icon" />
            GitHub Stats
          </a>
          <nav className="site-nav">
            {NAV_LINKS.map((link) => (
              <a key={link.href} href={link.href}>
                {link.label}
              </a>
            ))}
          </nav>
          <div className="site-header-actions">
            <ThemeToggle />
            <LinkButton
              href={SITE.repoUrl}
              target="_blank"
              rel="noreferrer"
              variant="ghost"
              icon={<Icon name="github-icon" />}
            >
              GitHub
            </LinkButton>
          </div>
        </div>
      </Container>
    </header>
  )
}
