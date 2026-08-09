export function normaliseTag(tag: string): string {
  return tag.trim().toLowerCase();
}

export function uniqueTags(tags: string[]): string[] {
  return [...new Set(tags.map(normaliseTag))];
}

export function addTag(tags: string[], tag: string): string[] {
  const norm = normaliseTag(tag);
  if (!norm) return tags;
  if (tags.map(normaliseTag).includes(norm)) return tags;
  return [...tags, norm];
}

export function removeTag(tags: string[], tag: string): string[] {
  const norm = normaliseTag(tag);
  return tags.filter((t) => normaliseTag(t) !== norm);
}

export function hasTag(tags: string[], tag: string): boolean {
  return tags.map(normaliseTag).includes(normaliseTag(tag));
}

export function intersectTags(lists: string[][]): string[] {
  if (lists.length === 0) return [];
  const [first, ...rest] = lists.map((l) => l.map(normaliseTag));
  return first.filter((t) => rest.every((l) => l.includes(t)));
}

export function unionTags(lists: string[][]): string[] {
  return uniqueTags(lists.flat());
}

export function tagFrequency(lists: string[][]): Record<string, number> {
  const freq: Record<string, number> = {};
  for (const list of lists) {
    for (const tag of list.map(normaliseTag)) {
      freq[tag] = (freq[tag] ?? 0) + 1;
    }
  }
  return freq;
}

export function topTags(lists: string[][], limit: number): string[] {
  const freq = tagFrequency(lists);
  return Object.entries(freq)
    .sort(([, a], [, b]) => b - a)
    .slice(0, limit)
    .map(([tag]) => tag);
}

export function renameTags(tags: string[], from: string, to: string): string[] {
  const normFrom = normaliseTag(from);
  return tags.map((t) => (normaliseTag(t) === normFrom ? normaliseTag(to) : t));
}

export function filterTagsByPrefix(tags: string[], prefix: string): string[] {
  const normPrefix = normaliseTag(prefix);
  return tags.filter((t) => normaliseTag(t).startsWith(normPrefix));
}
