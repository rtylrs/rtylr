/** Anchor link helpers for the docs app. */

/**
 * Converts a heading text to a URL-safe anchor ID.
 */
export function headingToAnchor(heading: string): string {
  return heading
    .trim()
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * Returns the full hash URL for a given anchor on a page.
 */
export function anchorUrl(pageUrl: string, anchor: string): string {
  const hash = anchor.startsWith('#') ? anchor : `#${anchor}`;
  return `${pageUrl}${hash}`;
}

/**
 * Extracts the anchor fragment from a URL (the part after #).
 */
export function extractAnchor(url: string): string {
  const idx = url.indexOf('#');
  return idx >= 0 ? url.slice(idx + 1) : '';
}

/**
 * Returns true if the URL contains an anchor fragment.
 */
export function hasAnchor(url: string): boolean {
  return url.includes('#');
}

/**
 * Returns true if two anchors point to the same heading (case-insensitive).
 */
export function anchorsMatch(a: string, b: string): boolean {
  return a.replace(/^#/, '').toLowerCase() === b.replace(/^#/, '').toLowerCase();
}
