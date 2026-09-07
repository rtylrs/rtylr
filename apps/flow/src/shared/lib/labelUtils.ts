import type { Label } from '../stores/flowStore';

/** Find a label by its ID from a label list. */
export function findLabelById(labels: Label[], id: string): Label | undefined {
  return labels.find((l) => l.id === id);
}

/** Resolve label IDs to label objects; silently drops missing IDs. */
export function resolveLabels(labels: Label[], ids: string[]): Label[] {
  return ids.flatMap((id) => {
    const label = findLabelById(labels, id);
    return label ? [label] : [];
  });
}

/** Sort labels alphabetically by name. */
export function sortLabelsAlpha(labels: Label[]): Label[] {
  return [...labels].sort((a, b) => a.name.localeCompare(b.name));
}

/** Filter labels whose name contains the search query (case-insensitive). */
export function filterLabels(labels: Label[], query: string): Label[] {
  const q = query.trim().toLowerCase();
  if (!q) return labels;
  return labels.filter((l) => l.name.toLowerCase().includes(q));
}
