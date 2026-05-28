import { describe, it, expect } from 'vitest';
import { searchMenuItems, itemMatchesQuery } from './searchMenuItems';

const items = [
  { name: 'Cheese Burger', sku: 'BRG-001' },
  { name: 'Veggie Wrap', sku: 'WRP-002' },
  { name: 'French Fries', sku: 'FRY-003' },
  { name: 'Chocolate Shake', sku: 'SHK-004' },
  { name: 'Caesar Salad' },
];

describe('searchMenuItems', () => {
  it('returns all items when query is empty', () => {
    expect(searchMenuItems(items, '')).toHaveLength(items.length);
  });

  it('returns all items when query is only whitespace', () => {
    expect(searchMenuItems(items, '   ')).toHaveLength(items.length);
  });

  it('filters by name substring (case-insensitive)', () => {
    const result = searchMenuItems(items, 'burger');
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Cheese Burger');
  });

  it('filters by SKU substring', () => {
    const result = searchMenuItems(items, 'WRP');
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Veggie Wrap');
  });

  it('returns empty array when no items match', () => {
    expect(searchMenuItems(items, 'pizza')).toHaveLength(0);
  });

  it('works for items without a SKU field', () => {
    const result = searchMenuItems(items, 'caesar');
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Caesar Salad');
  });
});

describe('itemMatchesQuery', () => {
  it('returns true for an empty query', () => {
    expect(itemMatchesQuery({ name: 'Burger' }, '')).toBe(true);
  });

  it('returns true when name matches', () => {
    expect(itemMatchesQuery({ name: 'Cheese Burger' }, 'cheese')).toBe(true);
  });

  it('returns true when SKU matches', () => {
    expect(itemMatchesQuery({ name: 'Burger', sku: 'BRG-001' }, 'BRG')).toBe(true);
  });

  it('returns false when neither name nor SKU matches', () => {
    expect(itemMatchesQuery({ name: 'Burger', sku: 'BRG-001' }, 'xyz')).toBe(false);
  });

  it('is case-insensitive for name comparison', () => {
    expect(itemMatchesQuery({ name: 'BURGER' }, 'burger')).toBe(true);
  });
});
