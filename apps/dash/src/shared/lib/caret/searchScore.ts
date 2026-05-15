/**
 * Score a chat against a query: higher = more relevant. Used by the
 * history search list to order results when results are heterogeneous
 * across titles and summaries.
 */

interface Scorable {
  title: string;
  summary?: string;
}

export function scoreChat(c: Scorable, query: string): number {
  const q = query.toLowerCase().trim();
  if (!q) return 0;
  const t = c.title.toLowerCase();
  const s = (c.summary ?? '').toLowerCase();

  let score = 0;
  if (t === q) score += 10;
  if (t.startsWith(q)) score += 6;
  if (t.includes(q)) score += 3;
  if (s.includes(q)) score += 1.5;
  // Words in query that appear in title boost score per-word.
  const words = q.split(/\s+/).filter(Boolean);
  for (const w of words) {
    if (t.includes(w)) score += 0.7;
    if (s.includes(w)) score += 0.3;
  }
  return score;
}
