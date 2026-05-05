// Hex extraídos de apps/web/app/assets/css/main.css — fonte de verdade.
export const colors = {
  background: '#121212',
  card: '#1a1a1a',
  sidebar: '#161616',
  primary: '#ffb700',
  primaryFg: '#121212',
  foreground: '#e1e1e1',
  mutedFg: '#888888',
  border: '#292929',
  input: '#292929',
  secondary: '#242424',
  destructive: '#cc3030',
  online: '#22c55e',
  offline: '#71717a',
  overlay: 'rgba(18,18,18,0.7)',
} as const;

export const radius = {
  sm: 4,
  md: 6,
  lg: 10,
  xl: 16,
  full: 9999,
} as const;

export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  xxl: 32,
} as const;

export type ColorKey = keyof typeof colors;
