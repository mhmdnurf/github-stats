import type { ReactNode } from 'react'

export function Badge({ children, active = false }: { children: ReactNode; active?: boolean }) {
  return <span className={`badge${active ? ' badge-active' : ''}`}>{children}</span>
}
