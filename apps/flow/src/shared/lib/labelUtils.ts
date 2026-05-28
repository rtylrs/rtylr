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

/** Returns true if a hex color is considered "dark" (useful for text contrast). */
export function isColorDark(hex: string): boolean {
  const clean = hex.replace('#', '');
  const r = parseInt(clean.slice(0, 2), 16);
  const g = parseInt(clean.slice(2, 4), 16);
  const b = parseInt(clean.slice(4, 6), 16);
  // Standard luminance formula
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return luminance < 0.5;
}

/** Returns '#ffffff' or '#000000' depending on the background color for contrast. */
export function labelTextColor(bgHex: string): '#ffffff' | '#000000' {
  return isColorDark(bgHex) ? '#ffffff' : '#000000';
}

/** Filter labels whose name contains the search query (case-insensitive). */
export function filterLabels(labels: Label[], query: string): Label[] {
  const q = query.trim().toLowerCase();
  if (!q) return labels;
  return labels.filter((l) => l.name.toLowerCase().includes(q));
}
