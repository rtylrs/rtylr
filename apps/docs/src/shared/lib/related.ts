/** Related and random document selection helpers for the docs app. */

export interface DocItem {
  slug: string;
  title?: string;
  tags?: string[];
}

/**
 * Returns documents related to the given doc by shared tags.
 * The current doc is excluded. Results are sorted by tag overlap (descending).
 */
export function getRelated(current: DocItem, all: DocItem[], maxResults = 3): DocItem[] {
  const currentTags = new Set((current.tags ?? []).map((t) => t.toLowerCase()));
  if (currentTags.size === 0) return [];

  return all
    .filter((doc) => doc.slug !== current.slug)
    .map((doc) => {
      const overlap = (doc.tags ?? []).filter((t) => currentTags.has(t.toLowerCase())).length;
      return { doc, overlap };
    })
    .filter(({ overlap }) => overlap > 0)
    .sort((a, b) => b.overlap - a.overlap)
    .slice(0, maxResults)
    .map(({ doc }) => doc);
}

/**
 * Returns up to `count` randomly-selected documents, excluding the given slug.
 */
export function getRandomDocs(docs: DocItem[], excludeSlug: string, count = 3): DocItem[] {
  const pool = docs.filter((d) => d.slug !== excludeSlug);
  const shuffled = [...pool].sort(() => Math.random() - 0.5);
  return shuffled.slice(0, count);
}
