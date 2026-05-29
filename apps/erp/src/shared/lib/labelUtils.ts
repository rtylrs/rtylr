/**
 * Product label and tag formatting helpers for the ERP.
 * All functions are pure with no side-effects.
 */

export type LabelVariant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info';

export interface ProductLabel {
  text: string;
  variant: LabelVariant;
}

/** Normalise a tag string: lowercase, trim, replace spaces with hyphens. */
export function normalizeTag(tag: string): string {
  return tag.trim().toLowerCase().replace(/\s+/g, '-');
}

/** Remove duplicate tags (case-insensitive). */
export function deduplicateTags(tags: string[]): string[] {
  const seen = new Set<string>();
  const result: string[] = [];
  for (const tag of tags) {
    const key = tag.toLowerCase().trim();
    if (!seen.has(key)) {
      seen.add(key);
      result.push(tag);
    }
  }
  return result;
}

/** Return a label for a product stock status. */
export function stockStatusLabel(stock: number, minStock: number): ProductLabel {
  if (stock === 0) return { text: 'Out of Stock', variant: 'danger' };
  if (stock <= minStock) return { text: 'Low Stock', variant: 'warning' };
  return { text: 'In Stock', variant: 'success' };
}

/** Return a label for a purchase order status. */
export function poStatusLabel(status: string): ProductLabel {
  const map: Record<string, ProductLabel> = {
    Draft: { text: 'Draft', variant: 'default' },
    Sent: { text: 'Sent', variant: 'primary' },
    Received: { text: 'Received', variant: 'success' },
    'Pending Approval': { text: 'Pending Approval', variant: 'warning' },
    Approved: { text: 'Approved', variant: 'success' },
  };
  return map[status] ?? { text: status, variant: 'default' };
}

/** Sort tags alphabetically. */
export function sortTags(tags: string[]): string[] {
  return [...tags].sort((a, b) => a.localeCompare(b));
}

/** Parse a comma-separated tag string into a cleaned array. */
export function parseTags(raw: string): string[] {
  return raw
    .split(',')
    .map((t) => t.trim())
    .filter(Boolean);
}

/** Serialize a tag array to a comma-separated string. */
export function serializeTags(tags: string[]): string {
  return tags.join(', ');
}
