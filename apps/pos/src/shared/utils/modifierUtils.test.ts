import { describe, it, expect } from 'vitest';
import {
  isValidSelection,
  isRequired,
  isMultiSelect,
  defaultSelections,
  selectedPriceDelta,
  mostExpensiveModifier,
  sortModifiersByName,
  modifierSummary,
  isModifierSelected,
} from './modifierUtils';
import type { Modifier, ModifierGroup } from '../types/pos';

const makeModifier = (overrides: Partial<Modifier> = {}): Modifier => ({
  id: 'm1',
  name: 'Cheese',
  priceDelta: 0.5,
  ...overrides,
});

const makeGroup = (overrides: Partial<ModifierGroup> = {}): ModifierGroup => ({
  id: 'g1',
  name: 'Toppings',
  min: 0,
  max: 3,
  modifiers: [makeModifier()],
  ...overrides,
});

describe('isValidSelection', () => {
  it('returns true when selectedCount is within range', () => {
    expect(isValidSelection(makeGroup({ min: 1, max: 3 }), 2)).toBe(true);
  });

  it('returns true at exactly min', () => {
    expect(isValidSelection(makeGroup({ min: 1, max: 3 }), 1)).toBe(true);
  });

  it('returns true at exactly max', () => {
    expect(isValidSelection(makeGroup({ min: 1, max: 3 }), 3)).toBe(true);
  });

  it('returns false when below min', () => {
    expect(isValidSelection(makeGroup({ min: 1, max: 3 }), 0)).toBe(false);
  });

  it('returns false when above max', () => {
    expect(isValidSelection(makeGroup({ min: 1, max: 3 }), 4)).toBe(false);
  });
});

describe('isRequired', () => {
  it('returns true when min > 0', () => {
    expect(isRequired(makeGroup({ min: 1 }))).toBe(true);
  });

  it('returns false when min is 0', () => {
    expect(isRequired(makeGroup({ min: 0 }))).toBe(false);
  });
});

describe('isMultiSelect', () => {
  it('returns true when max > 1', () => {
    expect(isMultiSelect(makeGroup({ max: 2 }))).toBe(true);
  });

  it('returns false when max is 1', () => {
    expect(isMultiSelect(makeGroup({ max: 1 }))).toBe(false);
  });
});

describe('defaultSelections', () => {
  it('returns modifiers with defaultSelected set to true', () => {
    const group = makeGroup({
      modifiers: [
        makeModifier({ id: 'm1', defaultSelected: true }),
        makeModifier({ id: 'm2', defaultSelected: false }),
        makeModifier({ id: 'm3', defaultSelected: true }),
      ],
    });
    const result = defaultSelections(group);
    expect(result).toHaveLength(2);
    expect(result.map((m) => m.id)).toEqual(['m1', 'm3']);
  });

  it('returns an empty array when no modifiers are default-selected', () => {
    const group = makeGroup({ modifiers: [makeModifier({ defaultSelected: false })] });
    expect(defaultSelections(group)).toHaveLength(0);
  });
});

describe('selectedPriceDelta', () => {
  it('sums the price deltas of selected modifiers', () => {
    const selected = [
      makeModifier({ priceDelta: 0.5 }),
      makeModifier({ priceDelta: 1.0 }),
    ];
    expect(selectedPriceDelta(selected)).toBeCloseTo(1.5);
  });

  it('returns 0 for an empty selection', () => {
    expect(selectedPriceDelta([])).toBe(0);
  });

  it('ignores non-finite price deltas', () => {
    const selected = [makeModifier({ priceDelta: NaN }), makeModifier({ priceDelta: 1.0 })];
    expect(selectedPriceDelta(selected)).toBe(1.0);
  });
});

describe('mostExpensiveModifier', () => {
  it('returns the modifier with the highest price delta', () => {
    const group = makeGroup({
      modifiers: [
        makeModifier({ id: 'a', priceDelta: 0.5 }),
        makeModifier({ id: 'b', priceDelta: 2.0 }),
        makeModifier({ id: 'c', priceDelta: 1.0 }),
      ],
    });
    expect(mostExpensiveModifier(group)?.id).toBe('b');
  });

  it('returns undefined for a group with no modifiers', () => {
    expect(mostExpensiveModifier(makeGroup({ modifiers: [] }))).toBeUndefined();
  });
});

describe('sortModifiersByName', () => {
  it('sorts modifiers alphabetically by name', () => {
    const group = makeGroup({
      modifiers: [
        makeModifier({ id: 'b', name: 'Mayo' }),
        makeModifier({ id: 'a', name: 'Cheese' }),
      ],
    });
    const sorted = sortModifiersByName(group);
    expect(sorted[0].name).toBe('Cheese');
    expect(sorted[1].name).toBe('Mayo');
  });

  it('does not mutate the original modifiers array', () => {
    const group = makeGroup({
      modifiers: [
        makeModifier({ id: 'b', name: 'Mayo' }),
        makeModifier({ id: 'a', name: 'Cheese' }),
      ],
    });
    sortModifiersByName(group);
    expect(group.modifiers[0].name).toBe('Mayo');
  });
});

describe('modifierSummary', () => {
  it('returns a comma-separated list of modifier names', () => {
    const selected = [makeModifier({ name: 'Cheese' }), makeModifier({ name: 'Bacon' })];
    expect(modifierSummary(selected)).toBe('Cheese, Bacon');
  });

  it('returns an empty string for an empty selection', () => {
    expect(modifierSummary([])).toBe('');
  });
});

describe('isModifierSelected', () => {
  it('returns true when the modifier id is in the selection', () => {
    const selected = [makeModifier({ id: 'm1' })];
    expect(isModifierSelected(selected, 'm1')).toBe(true);
  });

  it('returns false when the modifier id is not in the selection', () => {
    const selected = [makeModifier({ id: 'm1' })];
    expect(isModifierSelected(selected, 'm2')).toBe(false);
  });
});
