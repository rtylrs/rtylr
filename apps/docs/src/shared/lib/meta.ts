// SEO meta tag helpers for the docs application.

const SITE_NAME = 'rtylr docs';
const BASE_URL = 'https://docs.rtylr.com';

/**
 * pageTitle returns the page title with the site name appended.
 */
export function pageTitle(title: string): string {
  if (!title.trim()) return SITE_NAME;
  return `${title} | ${SITE_NAME}`;
}

/**
 * truncateDescription truncates a description to the SEO-optimal length.
 */
export function truncateDescription(description: string, maxLength = 160): string {
  if (description.length <= maxLength) return description;
  return description.slice(0, maxLength - 3) + '...';
}

/**
 * canonicalUrl builds an absolute canonical URL from a path.
 */
export function canonicalUrl(path: string, base = BASE_URL): string {
  const normalised = path.startsWith('/') ? path : '/' + path;
  return base + normalised;
}

/**
 * normaliseTags normalises an array of tags by lowercasing, trimming, deduplicating, and sorting.
 */
export function normaliseTags(tags: string[]): string[] {
  return [...new Set(tags.map((t) => t.trim().toLowerCase()))].sort();
}

/**
 * sharesTags returns true when two tag arrays have at least one tag in common.
 */
export function sharesTags(a: string[], b: string[]): boolean {
  const bNorm = new Set(normaliseTags(b));
  return normaliseTags(a).some((tag) => bNorm.has(tag));
}
