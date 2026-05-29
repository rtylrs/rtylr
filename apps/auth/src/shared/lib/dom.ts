/**
 * Returns true when the code is running in a browser environment.
 */
export function isBrowser(): boolean {
  return typeof window !== 'undefined' && typeof document !== 'undefined';
}

/**
 * Returns the current page URL's origin, or an empty string when not in a browser.
 */
export function getCurrentOrigin(): string {
  if (!isBrowser()) return '';
  return window.location.origin;
}

/**
 * Returns the current page URL's hostname, or an empty string when not in a browser.
 */
export function getCurrentHostname(): string {
  if (!isBrowser()) return '';
  return window.location.hostname;
}

/**
 * Sets the browser's document title.
 */
export function setDocumentTitle(title: string): void {
  if (!isBrowser()) return;
  document.title = title;
}

/**
 * Reads a query parameter from the current page URL.
 * Returns null when not in a browser or the param is absent.
 */
export function getQueryParam(name: string): string | null {
  if (!isBrowser()) return null;
  return new URLSearchParams(window.location.search).get(name);
}

/**
 * Returns true when the user prefers reduced motion.
 */
export function prefersReducedMotion(): boolean {
  if (!isBrowser()) return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Returns true when the user prefers a dark color scheme.
 */
export function prefersDarkMode(): boolean {
  if (!isBrowser()) return false;
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
}
