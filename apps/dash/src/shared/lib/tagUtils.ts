/**
 * Tag management helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export interface Tag {
  id: string;
  name: string;
  color?: string;
}

export function normalizeTagName(name: string): string {
  return name.trim().toLowerCase();
}

export function isValidTagName(name: string): boolean {
  const t = name.trim();
  return t.length > 0 && t.length <= 32 && /^[\w\s-]+$/.test(t);
}

export function tagExists(tags: Tag[], name: string): boolean {
  const n = normalizeTagName(name);
  return tags.some((t) => normalizeTagName(t.name) === n);
}

export function addTag(tags: Tag[], tag: Tag): Tag[] {
  if (tags.some((t) => t.id === tag.id)) return tags;
  return [...tags, tag];
}

export function removeTag(tags: Tag[], id: string): Tag[] {
  return tags.filter((t) => t.id !== id);
}

export function toggleTag(tags: Tag[], tag: Tag): Tag[] {
  return tags.some((t) => t.id === tag.id) ? removeTag(tags, tag.id) : addTag(tags, tag);
}

export function sortTags(tags: Tag[]): Tag[] {
  return [...tags].sort((a, b) => a.name.localeCompare(b.name));
}

export function filterTags(tags: Tag[], query: string): Tag[] {
  const q = query.trim().toLowerCase();
  if (!q) return tags;
  return tags.filter((t) => t.name.toLowerCase().includes(q));
}
