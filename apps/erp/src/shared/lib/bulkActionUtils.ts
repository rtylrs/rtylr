/**
 * Bulk selection and action helpers for ERP list views.
 * All functions are pure with no side-effects.
 */

export interface BulkSelection {
  selectedIds: Set<number>;
  totalCount: number;
}

/** Create an empty bulk selection. */
export function emptySelection(totalCount: number): BulkSelection {
  return { selectedIds: new Set(), totalCount };
}

/** Toggle selection of a single item. */
export function toggleItem(selection: BulkSelection, id: number): BulkSelection {
  const next = new Set(selection.selectedIds);
  if (next.has(id)) {
    next.delete(id);
  } else {
    next.add(id);
  }
  return { ...selection, selectedIds: next };
}

/** Select all items (given all ids). */
export function selectAll(selection: BulkSelection, allIds: number[]): BulkSelection {
  return { ...selection, selectedIds: new Set(allIds) };
}

/** Clear all selections. */
export function clearSelection(selection: BulkSelection): BulkSelection {
  return { ...selection, selectedIds: new Set() };
}

/** Return true if the item is selected. */
export function isSelected(selection: BulkSelection, id: number): boolean {
  return selection.selectedIds.has(id);
}

/** Return the count of selected items. */
export function selectedCount(selection: BulkSelection): number {
  return selection.selectedIds.size;
}

/** Return true if all items are selected. */
export function isAllSelected(selection: BulkSelection): boolean {
  return selection.selectedIds.size === selection.totalCount && selection.totalCount > 0;
}

/** Return true if some but not all items are selected. */
export function isPartiallySelected(selection: BulkSelection): boolean {
  return selection.selectedIds.size > 0 && selection.selectedIds.size < selection.totalCount;
}

/** Return selected ids as a sorted array. */
export function selectedIdsArray(selection: BulkSelection): number[] {
  return Array.from(selection.selectedIds).sort((a, b) => a - b);
}
