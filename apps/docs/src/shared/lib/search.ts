import type { DocContent } from '../content';

export interface SearchResult {
  slug: string;
  title: string;
  snippet: string;
}

/**
 * Strips HTML tags from a string, returning plain text.
 */
export function stripHtml(html: string): string {
  return html.replace(/<[^>]+>/g, '');
}

/**
 * Extracts a contextual snippet from body text around the first occurrence of query.
 * Returns at most (contextBefore + query.length + contextAfter) visible characters.
 */
export function extractSnippet(bodyText: string, query: string, contextChars = 60): string {
  const q = query.toLowerCase();
  const idx = bodyText.toLowerCase().indexOf(q);
  if (idx === -1) {
    return bodyText.slice(0, 120).trim() + (bodyText.length > 120 ? '...' : '');
  }
  const start = Math.max(0, idx - contextChars);
  const end = Math.min(bodyText.length, idx + query.length + contextChars);
  return (
    (start > 0 ? '...' : '') +
    bodyText.slice(start, end).trim() +
    (end < bodyText.length ? '...' : '')
  );
}

/**
 * Searches the documentation content registry for entries matching the query.
 * Matches against both titles and body text (with HTML stripped).
 * Returns an array of results sorted title-matches first.
 */
export function searchDocs(contentMap: Record<string, DocContent>, query: string): SearchResult[] {
  const trimmed = query.trim();
  if (!trimmed) return [];

  const q = trimmed.toLowerCase();
  const titleMatches: SearchResult[] = [];
  const bodyMatches: SearchResult[] = [];

  for (const [slug, doc] of Object.entries(contentMap)) {
    const titleLower = doc.title.toLowerCase();
    const bodyText = stripHtml(doc.body);
    const titleMatch = titleLower.includes(q);
    const bodyMatch = bodyText.toLowerCase().includes(q);

    if (!titleMatch && !bodyMatch) continue;

    const snippet = bodyMatch
      ? extractSnippet(bodyText, trimmed)
      : bodyText.slice(0, 120).trim() + '...';

    const result: SearchResult = { slug, title: doc.title, snippet };
    if (titleMatch) {
      titleMatches.push(result);
    } else {
      bodyMatches.push(result);
    }
  }

  return [...titleMatches, ...bodyMatches];
}
