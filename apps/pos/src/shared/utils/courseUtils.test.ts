import { describe, it, expect } from 'vitest';
import {
  hasCourse,
  itemsByCourse,
  filterByCourse,
  unassignedItems,
  courseLabel,
  activeCourses,
} from './courseUtils';
import type { OrderLineItem } from '../types/pos';

const makeItem = (overrides: Partial<OrderLineItem> = {}): OrderLineItem => ({
  id: 'i1',
  name: 'Burger',
  price: 10,
  quantity: 1,
  ...overrides,
});

describe('hasCourse', () => {
  it('returns true when a course is assigned', () => {
    expect(hasCourse(makeItem({ course: 'main' }))).toBe(true);
  });

  it('returns false when course is undefined', () => {
    expect(hasCourse(makeItem())).toBe(false);
  });
});

describe('itemsByCourse', () => {
  it('groups items into their respective course buckets', () => {
    const items = [
      makeItem({ id: 'a', course: 'main' }),
      makeItem({ id: 'b', course: 'drink' }),
      makeItem({ id: 'c' }),
    ];
    const grouped = itemsByCourse(items);
    expect(grouped.main).toHaveLength(1);
    expect(grouped.drink).toHaveLength(1);
    expect(grouped.unassigned).toHaveLength(1);
    expect(grouped.appetizer).toHaveLength(0);
  });
});

describe('filterByCourse', () => {
  it('returns only items for the given course', () => {
    const items = [
      makeItem({ id: 'a', course: 'main' }),
      makeItem({ id: 'b', course: 'dessert' }),
    ];
    const result = filterByCourse(items, 'main');
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe('a');
  });
});

describe('unassignedItems', () => {
  it('returns items without a course', () => {
    const items = [makeItem({ id: 'a', course: 'main' }), makeItem({ id: 'b' })];
    const result = unassignedItems(items);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe('b');
  });

  it('returns empty array when all items have a course', () => {
    expect(unassignedItems([makeItem({ course: 'main' })])).toHaveLength(0);
  });
});

describe('courseLabel', () => {
  it('returns the display label for each course', () => {
    expect(courseLabel('appetizer')).toBe('Appetizer');
    expect(courseLabel('main')).toBe('Main');
    expect(courseLabel('dessert')).toBe('Dessert');
    expect(courseLabel('drink')).toBe('Drink');
    expect(courseLabel('other')).toBe('Other');
  });
});

describe('activeCourses', () => {
  it('returns courses present in the item list in display order', () => {
    const items = [
      makeItem({ course: 'drink' }),
      makeItem({ course: 'main' }),
      makeItem({ course: 'appetizer' }),
    ];
    expect(activeCourses(items)).toEqual(['appetizer', 'main', 'drink']);
  });

  it('excludes courses not present in the items', () => {
    const items = [makeItem({ course: 'main' })];
    expect(activeCourses(items)).toEqual(['main']);
  });

  it('returns an empty array when no items have courses', () => {
    expect(activeCourses([makeItem()])).toEqual([]);
  });
});
