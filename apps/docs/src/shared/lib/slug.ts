/** Slug helpers for the docs app. */

/**
 * Converts a string to a URL-safe slug.
 */
export function toSlug(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * Converts a slug back to a human-readable title.
 */
export function slugToTitle(slug: string): string {
  return slug
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Returns true if the value is a valid slug (lowercase, hyphens, no leading/trailing hyphens).
 */
export function isValidSlug(value: string): boolean {
  return /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value);
}

/**
 * Returns a slug with a numeric suffix to make it unique among existing slugs.
 */
export function uniqueSlug(base: string, existing: string[]): string {
  const slug = toSlug(base);
  if (!existing.includes(slug)) return slug;
  let i = 2;
  while (existing.includes(`${slug}-${i}`)) i++;
  return `${slug}-${i}`;
}
