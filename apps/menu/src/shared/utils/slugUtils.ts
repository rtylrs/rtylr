/**
 * Converts a display name into a URL-safe slug.
 * Lowercases, replaces whitespace with hyphens, and removes non-alphanumeric chars.
 */
export function toSlug(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

/**
 * Returns true when a string is a valid slug (only lowercase letters, digits, and hyphens).
 */
export function isValidSlug(slug: string): boolean {
  return /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(slug);
}

/**
 * Extracts the business slug from a URL path like /b/:slug/... or /:slug.
 * Returns null when no slug is found.
 */
export function extractSlugFromPath(path: string): string | null {
  const match = path.match(/^\/b\/([a-z0-9-]+)/);
  if (match) return match[1];
  const simple = path.match(/^\/([a-z0-9-]+)/);
  return simple ? simple[1] : null;
}
