/**
 * Text highlight helpers for the docs app.
 */

/**
 * Wraps all occurrences of a term in a string with a marker (default `<mark>`).
 */
export function highlightTerm(text: string, term: string, tag = 'mark'): string {
  if (!term.trim()) return text;
  const escaped = term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const re = new RegExp(`(${escaped})`, 'gi');
  return text.replace(re, `<${tag}>$1</${tag}>`);
}

/**
 * Returns an excerpt from text centred around the first occurrence of a term.
 * Falls back to the beginning of the text when the term is not found.
 */
export function excerptAround(text: string, term: string, windowSize = 100): string {
  if (!term.trim()) return text;
  const idx = text.toLowerCase().indexOf(term.toLowerCase());
  if (idx === -1) return text;
  const start = Math.max(0, idx - windowSize);
  const end = Math.min(text.length, idx + term.length + windowSize);
  const excerpt = text.slice(start, end);
  const prefix = start > 0 ? '...' : '';
  const suffix = end < text.length ? '...' : '';
  return `${prefix}${excerpt}${suffix}`;
}

/**
 * Returns true when a string contains all the given terms (case-insensitive).
 */
export function containsAll(text: string, terms: string[]): boolean {
  const lower = text.toLowerCase();
  return terms.every((t) => lower.includes(t.toLowerCase()));
}

/**
 * Strips HTML tags from a string.
 */
export function stripTags(html: string): string {
  return html.replace(/<[^>]*>/g, '');
}
