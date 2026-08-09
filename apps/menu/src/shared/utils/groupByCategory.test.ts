import { describe, it, expect } from 'vitest';
import { groupBy, groupByCategory, categoryNames, itemsForCategory } from './groupByCategory';

const items = [
  { id: '1', name: 'Burger', category: 'mains' },
  { id: '2', name: 'Fries', category: 'sides' },
  { id: '3', name: 'Shake', category: 'mains' },
  { id: '4', name: 'Salad', category: 'sides' },
  { id: '5', name: 'Water', category: 'drinks' },
];

describe('groupBy', () => {
  it('groups items by a derived key', () => {
    const grouped = groupBy(items, (i) => i.category);
    expect(Object.keys(grouped).sort()).toEqual(['drinks', 'mains', 'sides']);
  });

  it('places each item under the correct key', () => {
    const grouped = groupBy(items, (i) => i.category);
    expect(grouped['mains']).toHaveLength(2);
    expect(grouped['sides']).toHaveLength(2);
    expect(grouped['drinks']).toHaveLength(1);
  });

  it('returns an empty object for an empty array', () => {
    expect(groupBy([], (i: { category: string }) => i.category)).toEqual({});
  });

  it('handles all items sharing the same key', () => {
    const same = [
      { id: '1', category: 'food' },
      { id: '2', category: 'food' },
    ];
    const grouped = groupBy(same, (i) => i.category);
    expect(grouped['food']).toHaveLength(2);
  });
});

describe('groupByCategory', () => {
  it('groups menu items by category field', () => {
    const grouped = groupByCategory(items);
    expect(grouped['mains']).toHaveLength(2);
    expect(grouped['sides']).toHaveLength(2);
  });

  it('places items with empty category under uncategorized', () => {
    const withEmpty = [{ id: '9', name: 'Mystery', category: '' }];
    const grouped = groupByCategory(withEmpty);
    expect(grouped['uncategorized']).toHaveLength(1);
  });

  it('returns an empty object for an empty array', () => {
    expect(groupByCategory([])).toEqual({});
  });
});

describe('categoryNames', () => {
  it('returns a sorted list of unique category names', () => {
    const names = categoryNames(items);
    expect(names).toEqual(['drinks', 'mains', 'sides']);
  });

  it('deduplicates repeated category values', () => {
    const dup = [{ category: 'mains' }, { category: 'mains' }, { category: 'sides' }];
    expect(categoryNames(dup)).toEqual(['mains', 'sides']);
  });

  it('uses uncategorized for empty category', () => {
    const names = categoryNames([{ category: '' }]);
    expect(names).toEqual(['uncategorized']);
  });
});

describe('itemsForCategory', () => {
  it('returns all items for a given category', () => {
    expect(itemsForCategory(items, 'mains')).toHaveLength(2);
  });

  it('is case-insensitive', () => {
    expect(itemsForCategory(items, 'MAINS')).toHaveLength(2);
  });

  it('returns an empty array when no items match', () => {
    expect(itemsForCategory(items, 'desserts')).toHaveLength(0);
  });
});
