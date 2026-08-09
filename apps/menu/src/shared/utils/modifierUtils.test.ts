import { describe, it, expect } from 'vitest';
import {
  modifierTotal,
  isModifierSelectionComplete,
  getGroupSelections,
  canSelectMore,
  formatModifierNames,
} from './modifierUtils';
import type { Modifier, ModifierGroup } from './modifierUtils';

const cheese: Modifier = { id: 1, name: 'Extra Cheese', priceDelta: 1.5 };
const bacon: Modifier = { id: 2, name: 'Bacon', priceDelta: 2.0 };
const lettuce: Modifier = { id: 3, name: 'Lettuce', priceDelta: 0 };

const toppingsGroup: ModifierGroup = {
  id: 10,
  name: 'Toppings',
  required: false,
  minSelections: 0,
  maxSelections: 2,
  modifiers: [cheese, bacon, lettuce],
};

const sauceGroup: ModifierGroup = {
  id: 20,
  name: 'Sauce',
  required: true,
  minSelections: 1,
  maxSelections: 1,
  modifiers: [
    { id: 4, name: 'Ketchup', priceDelta: 0 },
    { id: 5, name: 'Mustard', priceDelta: 0 },
  ],
};

describe('modifierTotal', () => {
  it('sums the price deltas of selected modifiers', () => {
    expect(modifierTotal([cheese, bacon])).toBe(3.5);
  });

  it('returns 0 for an empty selection', () => {
    expect(modifierTotal([])).toBe(0);
  });

  it('handles free modifiers correctly', () => {
    expect(modifierTotal([lettuce])).toBe(0);
  });
});

describe('isModifierSelectionComplete', () => {
  it('returns true when required group has a selection', () => {
    const selected = new Set([4]);
    expect(isModifierSelectionComplete([sauceGroup], selected)).toBe(true);
  });

  it('returns false when required group has no selection', () => {
    const selected = new Set<number>();
    expect(isModifierSelectionComplete([sauceGroup], selected)).toBe(false);
  });

  it('returns true for optional groups even with no selection', () => {
    const selected = new Set<number>();
    expect(isModifierSelectionComplete([toppingsGroup], selected)).toBe(true);
  });
});

describe('getGroupSelections', () => {
  it('returns only modifiers that are selected', () => {
    const selected = new Set([1, 3]);
    const result = getGroupSelections(toppingsGroup, selected);
    expect(result.map((m) => m.id)).toEqual([1, 3]);
  });

  it('returns empty when nothing is selected', () => {
    expect(getGroupSelections(toppingsGroup, new Set())).toHaveLength(0);
  });
});

describe('canSelectMore', () => {
  it('returns true when below the max', () => {
    const selected = new Set([1]);
    expect(canSelectMore(toppingsGroup, selected)).toBe(true);
  });

  it('returns false when at the max', () => {
    const selected = new Set([1, 2]);
    expect(canSelectMore(toppingsGroup, selected)).toBe(false);
  });
});

describe('formatModifierNames', () => {
  it('joins modifier names with a comma', () => {
    expect(formatModifierNames([cheese, bacon])).toBe('Extra Cheese, Bacon');
  });

  it('returns an empty string for no selections', () => {
    expect(formatModifierNames([])).toBe('');
  });
});
