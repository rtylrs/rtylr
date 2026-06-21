/** Caret design-token mirror for dash. The @voxire/caret package keeps
 *  these internal - duplicating the small set here avoids cross-package
 *  surface churn. Update the values together when the package's
 *  `ui/tokens` constants change. */
export const caretTokens = {
  accent: '#64c5b9',
  accentSoft: 'rgba(100, 197, 185, 0.14)',
  bg: '#141414',
  surface: '#1c1c1c',
  border: '#2a2a2a',
  text: '#ffffff',
  textMuted: '#a1a1aa',
  textDim: '#71717a',
} as const;

export const caretRadius = {
  sm: '6px',
  md: '8px',
  lg: '10px',
  xl: '14px',
} as const;
