import type { ReactNode } from 'react'
import { Container } from './Container'

type SectionProps = {
  id?: string
  eyebrow?: string
  title: ReactNode
  description?: ReactNode
  children?: ReactNode
}

export function Section({ id, eyebrow, title, description, children }: SectionProps) {
  return (
    <section id={id} className="section">
      <Container>
        <div className="section-head">
          {eyebrow ? <span className="eyebrow">{eyebrow}</span> : null}
          <h2>{title}</h2>
          {description ? <p className="section-desc">{description}</p> : null}
        </div>
        {children}
      </Container>
    </section>
  )
}
