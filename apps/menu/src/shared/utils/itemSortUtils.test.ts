import { describe, it, expect } from 'vitest';
import { sortItems, topItems, filterByPriceRange } from './itemSortUtils';

const items = [
  { name: 'Burger', price: 12.99, calories: 650, orderCount: 100 },
  { name: 'Salad', price: 8.5, calories: 200, orderCount: 50 },
  { name: 'Pizza', price: 15.0, calories: 900, orderCount: 200 },
  { name: 'Wrap', price: 10.0, calories: 450, orderCount: 75 },
];

describe('sortItems', () => {
  it('sorts by name ascending', () => {
    const result = sortItems(items, 'name', 'asc');
    expect(result[0].name).toBe('Burger');
    expect(result[result.length - 1].name).toBe('Wrap');
  });

  it('sorts by name descending', () => {
    const result = sortItems(items, 'name', 'desc');
    expect(result[0].name).toBe('Wrap');
  });

  it('sorts by price ascending', () => {
    const result = sortItems(items, 'price', 'asc');
    expect(result[0].price).toBe(8.5);
  });

  it('sorts by price descending', () => {
    const result = sortItems(items, 'price', 'desc');
    expect(result[0].price).toBe(15.0);
  });

  it('sorts by popular (highest orderCount first)', () => {
    const result = sortItems(items, 'popular', 'asc');
    expect(result[0].orderCount).toBe(200);
  });

  it('does not mutate the original array', () => {
    const original = [...items];
    sortItems(items, 'name');
    expect(items).toEqual(original);
  });
});

describe('topItems', () => {
  it('returns the top N items by order count', () => {
    const result = topItems(items, 2);
    expect(result).toHaveLength(2);
    expect(result[0].name).toBe('Pizza');
    expect(result[1].name).toBe('Burger');
  });

  it('returns all items when N exceeds count', () => {
    expect(topItems(items, 100)).toHaveLength(items.length);
  });
});

describe('filterByPriceRange', () => {
  it('returns items within the price range', () => {
    const result = filterByPriceRange(items, 10, 13);
    expect(result.map((i) => i.name).sort()).toEqual(['Burger', 'Wrap']);
  });

  it('includes items at the boundary', () => {
    const result = filterByPriceRange(items, 8.5, 8.5);
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Salad');
  });

  it('returns empty when no items are in range', () => {
    expect(filterByPriceRange(items, 100, 200)).toHaveLength(0);
  });
});
