import { describe, it, expect } from 'vitest';
import {
  emptySelection,
  toggleItem,
  selectAll,
  clearSelection,
  isSelected,
  selectedCount,
  isAllSelected,
  isPartiallySelected,
  selectedIdsArray,
} from './bulkActionUtils';

describe('emptySelection', () => {
  it('creates an empty selection with correct total', () => {
    const sel = emptySelection(10);
    expect(sel.selectedIds.size).toBe(0);
    expect(sel.totalCount).toBe(10);
  });
});

describe('toggleItem', () => {
  it('adds an item if not selected', () => {
    const sel = emptySelection(5);
    const next = toggleItem(sel, 1);
    expect(isSelected(next, 1)).toBe(true);
  });

  it('removes an item if already selected', () => {
    const sel = emptySelection(5);
    const withItem = toggleItem(sel, 1);
    const withoutItem = toggleItem(withItem, 1);
    expect(isSelected(withoutItem, 1)).toBe(false);
  });

  it('does not mutate original', () => {
    const sel = emptySelection(5);
    toggleItem(sel, 1);
    expect(sel.selectedIds.size).toBe(0);
  });
});

describe('selectAll', () => {
  it('selects all provided ids', () => {
    const sel = emptySelection(3);
    const next = selectAll(sel, [1, 2, 3]);
    expect(isAllSelected(next)).toBe(true);
    expect(selectedCount(next)).toBe(3);
  });
});

describe('clearSelection', () => {
  it('clears all selected ids', () => {
    const sel = selectAll(emptySelection(3), [1, 2, 3]);
    const cleared = clearSelection(sel);
    expect(selectedCount(cleared)).toBe(0);
  });
});

describe('isAllSelected', () => {
  it('returns true when all items selected', () => {
    const sel = selectAll(emptySelection(2), [1, 2]);
    expect(isAllSelected(sel)).toBe(true);
  });

  it('returns false for empty selection', () => {
    expect(isAllSelected(emptySelection(5))).toBe(false);
  });

  it('returns false when total is 0', () => {
    expect(isAllSelected(emptySelection(0))).toBe(false);
  });
});

describe('isPartiallySelected', () => {
  it('returns true when some but not all selected', () => {
    const sel = toggleItem(emptySelection(3), 1);
    expect(isPartiallySelected(sel)).toBe(true);
  });

  it('returns false when none selected', () => {
    expect(isPartiallySelected(emptySelection(3))).toBe(false);
  });

  it('returns false when all selected', () => {
    const sel = selectAll(emptySelection(2), [1, 2]);
    expect(isPartiallySelected(sel)).toBe(false);
  });
});

describe('selectedIdsArray', () => {
  it('returns sorted array of selected ids', () => {
    let sel = emptySelection(5);
    sel = toggleItem(sel, 3);
    sel = toggleItem(sel, 1);
    sel = toggleItem(sel, 5);
    expect(selectedIdsArray(sel)).toEqual([1, 3, 5]);
  });
});
