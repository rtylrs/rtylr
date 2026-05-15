/**
 * Strip basic markdown syntax to produce a clean plain-text preview suitable
 * for chat list summaries (no rendering, no formatting).
 */
export function markdownToPreview(src: string, maxLen = 80): string {
  if (!src) return '';
  let s = src;
  // Code fences
  s = s.replace(/```[\s\S]*?```/g, ' ');
  // Inline code
  s = s.replace(/`([^`]+)`/g, '$1');
  // Headings / quotes / list markers
  s = s.replace(/^\s{0,3}(#{1,6}\s+|>\s+|[-*+]\s+|\d+\.\s+)/gm, '');
  // Links: keep label
  s = s.replace(/\[([^\]]+)\]\([^)]+\)/g, '$1');
  // Emphasis
  s = s.replace(/(\*\*|__)([^*_]+)\1/g, '$2');
  s = s.replace(/(\*|_)([^*_]+)\1/g, '$2');
  s = s.replace(/\s+/g, ' ').trim();
  if (s.length > maxLen) s = s.slice(0, maxLen - 1).trimEnd() + '…';
  return s;
}
