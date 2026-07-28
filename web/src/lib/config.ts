export const SITE = {
  repoUrl: 'https://github.com/mhmdnurf/github-stats',
  liveDemoUrl: 'https://github-stats-66745590752.asia-southeast2.run.app',
} as const

export const THEMES = ['default', 'light', 'dracula', 'tokyonight', 'gruvbox'] as const

export type ThemeName = (typeof THEMES)[number]

export function statsCardUrl(theme: ThemeName) {
  return `${SITE.liveDemoUrl}/stats?theme=${theme}`
}

export function languagesCardUrl(theme: ThemeName) {
  return `${SITE.liveDemoUrl}/languages?theme=${theme}`
}
