import { createContext } from 'react'

export type Mode = 'light' | 'dark'

export type ColorModeContextValue = { mode: Mode; toggle: () => void }

export const ColorModeContext = createContext<ColorModeContextValue | null>(null)
