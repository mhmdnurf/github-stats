import { Container } from '../ui/Container'
import { Icon } from '../ui/Icon'
import { SITE } from '../../lib/config'

const SOCIAL_LINKS = [
  { href: SITE.repoUrl, icon: 'github-icon', label: 'GitHub' },
]

export function Footer() {
  return (
    <footer className="site-footer">
      <Container>
        <div className="site-footer-inner">
          <span>GitHub Stats &mdash; self-hosted GitHub profile cards.</span>
          <ul className="footer-links">
            {SOCIAL_LINKS.map((link) => (
              <li key={link.href}>
                <a href={link.href} target="_blank" rel="noreferrer">
                  <Icon name={link.icon} className="button-icon" />
                  {link.label}
                </a>
              </li>
            ))}
          </ul>
        </div>
      </Container>
    </footer>
  )
}
