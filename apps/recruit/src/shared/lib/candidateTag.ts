// Candidate tag helpers for the recruit admin panel.

export interface CandidateTag {
  id: number;
  label: string;
  color: string;
}

/**
 * uniqueTags returns unique tags by label (case-insensitive).
 */
export function uniqueTags(tags: CandidateTag[]): CandidateTag[] {
  const seen = new Set<string>();
  return tags.filter((t) => {
    const key = t.label.toLowerCase();
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

/**
 * tagsByColor groups tags by their color value.
 */
export function tagsByColor(tags: CandidateTag[]): Record<string, CandidateTag[]> {
  return tags.reduce<Record<string, CandidateTag[]>>((acc, tag) => {
    const group = acc[tag.color] ?? [];
    acc[tag.color] = [...group, tag];
    return acc;
  }, {});
}

/**
 * hasTag returns true when the candidate has a tag with the given label.
 */
export function hasTag(tags: CandidateTag[], label: string): boolean {
  const lower = label.toLowerCase();
  return tags.some((t) => t.label.toLowerCase() === lower);
}

/**
 * sortTagsAlpha returns tags sorted alphabetically by label.
 */
export function sortTagsAlpha(tags: CandidateTag[]): CandidateTag[] {
  return [...tags].sort((a, b) => a.label.localeCompare(b.label));
}

/**
 * tagLabels extracts just the label strings from a tag array.
 */
export function tagLabels(tags: CandidateTag[]): string[] {
  return tags.map((t) => t.label);
}
