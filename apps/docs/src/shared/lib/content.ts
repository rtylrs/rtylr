import type { DocContent } from '../content';

/**
 * Converts a slug to a human-readable title.
 */
export function slugToTitle(slug: string): string {
  return slug
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Converts a title to a URL-safe slug.
 */
export function titleToSlug(title: string): string {
  return title
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

/**
 * Returns all slugs in the content map sorted alphabetically.
 */
export function getSortedSlugs(contentMap: Record<string, DocContent>): string[] {
  return Object.keys(contentMap).sort();
}

/**
 * Returns true when the given slug exists in the content map.
 */
export function hasSlug(contentMap: Record<string, DocContent>, slug: string): boolean {
  return Object.prototype.hasOwnProperty.call(contentMap, slug);
}

/**
 * Returns the plain-text length of a document's body (HTML tags stripped).
 */
export function bodyLength(doc: DocContent): number {
  return doc.body.replace(/<[^>]+>/g, '').length;
}

/**
 * Returns a list of slugs whose body or title contains the given keyword.
 */
export function slugsContaining(contentMap: Record<string, DocContent>, keyword: string): string[] {
  const q = keyword.toLowerCase();
  return Object.entries(contentMap)
    .filter(([, doc]) => {
      const bodyText = doc.body.replace(/<[^>]+>/g, '').toLowerCase();
      return doc.title.toLowerCase().includes(q) || bodyText.includes(q);
    })
    .map(([slug]) => slug);
}
