/**
 * Filtering helpers for docs content.
 */

export interface DocItem {
  title: string;
  tags?: string[];
  publishedAt?: string;
}

/**
 * Filters items by a search query matching title (case-insensitive).
 */
export function filterByQuery<T extends Pick<DocItem, 'title' | 'tags'>>(items: T[], query: string): T[] {
  const q = query.toLowerCase().trim();
  if (!q) return items;
  return items.filter(
    (item) =>
      item.title.toLowerCase().includes(q) ||
      (item.tags ?? []).some((tag) => tag.toLowerCase().includes(q))
  );
}

/**
 * Filters items that include ALL of the given tags (case-insensitive).
 */
export function filterByTags<T extends Pick<DocItem, 'tags'>>(items: T[], tags: string[]): T[] {
  if (tags.length === 0) return items;
  const lower = tags.map((t) => t.toLowerCase());
  return items.filter((item) =>
    lower.every((tag) => (item.tags ?? []).map((t) => t.toLowerCase()).includes(tag))
  );
}

/**
 * Filters items published on or after the given date string (YYYY-MM-DD).
 */
export function filterSince<T extends Pick<DocItem, 'publishedAt'>>(
  items: T[],
  since: string
): T[] {
  const sinceMs = new Date(since).getTime();
  return items.filter((item) => {
    if (!item.publishedAt) return false;
    return new Date(item.publishedAt).getTime() >= sinceMs;
  });
}

/**
 * Returns items that do NOT appear in the excluded slug set.
 */
export function excludeSlugs<T extends { slug: string }>(items: T[], slugs: string[]): T[] {
  const set = new Set(slugs);
  return items.filter((item) => !set.has(item.slug));
}
