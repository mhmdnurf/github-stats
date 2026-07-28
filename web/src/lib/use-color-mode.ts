import { useContext } from 'react'
import { ColorModeContext } from './color-mode-context'

export function useColorMode() {
  const ctx = useContext(ColorModeContext)
  if (!ctx) throw new Error('useColorMode must be used within a ColorModeProvider')
  return ctx
}
