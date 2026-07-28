import { useEffect, useState, type ReactNode } from 'react'
import { ColorModeContext, type Mode } from './color-mode-context'

function getInitialMode(): Mode {
  const stored = localStorage.getItem('color-mode')
  if (stored === 'light' || stored === 'dark') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function ColorModeProvider({ children }: { children: ReactNode }) {
  const [mode, setMode] = useState<Mode>(getInitialMode)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', mode)
    localStorage.setItem('color-mode', mode)
  }, [mode])

  return (
    <ColorModeContext.Provider
      value={{ mode, toggle: () => setMode((current) => (current === 'dark' ? 'light' : 'dark')) }}
    >
      {children}
    </ColorModeContext.Provider>
  )
}
