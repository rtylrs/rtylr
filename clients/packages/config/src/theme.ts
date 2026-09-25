/**
 * App theme tokens. Apply via injectTheme() in your app root,
 * or use the CSS variables directly (--primary-color, --secondary-color, etc.)
 */
export const theme = {
  palette: {
    primary: { main: '#151515' },
    secondary: { main: '#dc004e' },
    success: { main: '#2e7d32' },
    warning: { main: '#ed6c02' },
    error: { main: '#d32f2f' },
    info: { main: '#0288d1' },
  },
  typography: {
    fontFamily: '"Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", sans-serif',
  },
  shape: { borderRadius: 8 },
} as const;

/** Call once in your app root to apply theme as CSS variables */
export function injectTheme(): void {
  const root = document.documentElement;
  root.style.setProperty('--primary-color', theme.palette.primary.main);
  root.style.setProperty('--secondary-color', theme.palette.secondary.main);
  root.style.setProperty('--success-color', theme.palette.success.main);
  root.style.setProperty('--warning-color', theme.palette.warning.main);
  root.style.setProperty('--error-color', theme.palette.error.main);
  root.style.setProperty('--info-color', theme.palette.info.main);
  root.style.setProperty('--font-family', theme.typography.fontFamily);
}

export default theme;
