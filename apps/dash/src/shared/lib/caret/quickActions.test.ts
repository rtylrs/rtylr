import { describe, expect, it } from 'vitest';
import { QUICK_ACTIONS, quickActionsByScope } from './quickActions';

describe('QUICK_ACTIONS', () => {
  it('exposes a non-empty seed list', () => {
    expect(QUICK_ACTIONS.length).toBeGreaterThan(0);
  });

  it('uses unique ids', () => {
    const ids = QUICK_ACTIONS.map((q) => q.id);
    expect(new Set(ids).size).toBe(ids.length);
  });
});

describe('quickActionsByScope', () => {
  it('filters to the requested scope', () => {
    const inventory = quickActionsByScope('inventory');
    expect(inventory.every((q) => q.scope === 'inventory')).toBe(true);
  });
});
