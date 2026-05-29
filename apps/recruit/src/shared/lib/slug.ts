// Slug helpers for the recruit careers public URL module.

/**
 * toSlug converts a plain-text string to a URL-safe slug.
 */
export function toSlug(value: string): string {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * isValidSlug returns true when the string matches the slug format.
 */
export function isValidSlug(slug: string): boolean {
  return /^[a-z0-9]+(-[a-z0-9]+)*$/.test(slug);
}

/**
 * slugToTitle converts a slug back to a title-cased string.
 */
export function slugToTitle(slug: string): string {
  return slug
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * careerUrl builds the public career listing URL.
 */
export function careerUrl(companySlug: string, careerSlug: string, base = ''): string {
  return `${base}/${companySlug}/careers/${careerSlug}`;
}

/**
 * applyUrl builds the application form URL for a career.
 */
export function applyUrl(companySlug: string, careerSlug: string, base = ''): string {
  return `${base}/${companySlug}/careers/${careerSlug}/apply`;
}
