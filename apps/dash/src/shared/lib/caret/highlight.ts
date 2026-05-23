/**
 * Wrap occurrences of `query` inside `text` with <mark> tags for the chat
 * history search results. Case-insensitive, doesn't escape input - feed
 * already-escaped text.
 */
export function highlight(text: string, query: string): string {
  if (!query.trim()) return text;
  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return text.replace(new RegExp(`(${escaped})`, 'gi'), '<mark>$1</mark>');
}
