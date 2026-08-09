/** Media query and viewport helpers for the auth app. */

export type Breakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl';

const BREAKPOINTS: Record<Breakpoint, number> = {
  xs: 0,
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,
};

/**
 * Returns the current viewport width in pixels.
 * Returns 0 in non-browser environments.
 */
export function viewportWidth(): number {
  if (typeof window === 'undefined') return 0;
  return window.innerWidth;
}

/**
 * Returns true if the viewport width is at least the given breakpoint.
 */
export function isAtLeast(breakpoint: Breakpoint): boolean {
  return viewportWidth() >= BREAKPOINTS[breakpoint];
}

/**
 * Returns the current active breakpoint name.
 */
export function activeBreakpoint(): Breakpoint {
  const width = viewportWidth();
  const bps: Breakpoint[] = ['2xl', 'xl', 'lg', 'md', 'sm', 'xs'];
  for (const bp of bps) {
    if (width >= BREAKPOINTS[bp]) return bp;
  }
  return 'xs';
}

/**
 * Returns true if the user prefers reduced motion.
 */
export function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Returns true if the user prefers dark color scheme.
 */
export function prefersDarkMode(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
}
