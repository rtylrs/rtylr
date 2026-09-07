import { describe, it, expect } from 'vitest';

import { findLabelById, resolveLabels, sortLabelsAlpha, filterLabels } from './labelUtils';
import type { Label } from '../stores/flowStore';

function makeLabel(overrides: Partial<Label>): Label {
  return {
    id: 'l1',
    name: 'Bug',
    color: '#ef4444',
    ...overrides,
  } as Label;
}

describe('findLabelById', () => {
  it('finds a label by id', () => {
    const labels = [makeLabel({ id: 'l1' }), makeLabel({ id: 'l2', name: 'Feature' })];
    expect(findLabelById(labels, 'l2')?.name).toBe('Feature');
  });

  it('returns undefined for missing id', () => {
    expect(findLabelById([], 'l1')).toBeUndefined();
  });
});

describe('resolveLabels', () => {
  it('resolves ids to label objects', () => {
    const labels = [makeLabel({ id: 'l1', name: 'Bug' }), makeLabel({ id: 'l2', name: 'Feature' })];
    const resolved = resolveLabels(labels, ['l1', 'l2']);
    expect(resolved.map((l) => l.name)).toEqual(['Bug', 'Feature']);
  });

  it('silently drops missing ids', () => {
    const labels = [makeLabel({ id: 'l1' })];
    const resolved = resolveLabels(labels, ['l1', 'l99']);
    expect(resolved).toHaveLength(1);
  });

  it('returns empty array for empty ids', () => {
    expect(resolveLabels([makeLabel({})], [])).toEqual([]);
  });
});

describe('sortLabelsAlpha', () => {
  it('sorts labels by name alphabetically', () => {
    const labels = [
      makeLabel({ name: 'Zeta' }),
      makeLabel({ name: 'Alpha' }),
      makeLabel({ name: 'Mango' }),
    ];
    const sorted = sortLabelsAlpha(labels);
    expect(sorted.map((l) => l.name)).toEqual(['Alpha', 'Mango', 'Zeta']);
  });

  it('does not mutate the original array', () => {
    const labels = [makeLabel({ name: 'Z' }), makeLabel({ name: 'A' })];
    const original = [...labels];
    sortLabelsAlpha(labels);
    expect(labels).toEqual(original);
  });
});

describe('filterLabels', () => {
  it('filters by name', () => {
    const labels = [makeLabel({ name: 'Bug' }), makeLabel({ name: 'Feature' })];
    expect(filterLabels(labels, 'bug')).toHaveLength(1);
    expect(filterLabels(labels, 'bug')[0].name).toBe('Bug');
  });

  it('returns all for empty query', () => {
    const labels = [makeLabel({ name: 'Bug' }), makeLabel({ name: 'Feature' })];
    expect(filterLabels(labels, '')).toHaveLength(2);
  });

  it('is case-insensitive', () => {
    expect(filterLabels([makeLabel({ name: 'Bug' })], 'BUG')).toHaveLength(1);
  });
});
