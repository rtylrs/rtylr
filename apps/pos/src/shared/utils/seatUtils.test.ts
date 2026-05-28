import { describe, it, expect } from 'vitest';
import {
  hasSeat,
  itemsBySeat,
  occupiedSeats,
  groupBySeat,
  seatSubtotal,
  unassignedSeatItems,
} from './seatUtils';
import type { OrderLineItem } from '../types/pos';

const makeItem = (overrides: Partial<OrderLineItem> = {}): OrderLineItem => ({
  id: 'i1',
  name: 'Burger',
  price: 10,
  quantity: 1,
  ...overrides,
});

describe('hasSeat', () => {
  it('returns true when seat is a positive integer', () => {
    expect(hasSeat(makeItem({ seat: 2 }))).toBe(true);
  });

  it('returns false when seat is undefined', () => {
    expect(hasSeat(makeItem())).toBe(false);
  });

  it('returns false when seat is 0', () => {
    expect(hasSeat(makeItem({ seat: 0 }))).toBe(false);
  });
});

describe('itemsBySeat', () => {
  it('returns items assigned to the given seat', () => {
    const items = [makeItem({ id: 'a', seat: 1 }), makeItem({ id: 'b', seat: 2 })];
    expect(itemsBySeat(items, 1)).toHaveLength(1);
    expect(itemsBySeat(items, 1)[0].id).toBe('a');
  });

  it('returns an empty array for a seat with no items', () => {
    expect(itemsBySeat([makeItem({ seat: 1 })], 3)).toHaveLength(0);
  });
});

describe('occupiedSeats', () => {
  it('returns unique seat numbers in ascending order', () => {
    const items = [makeItem({ seat: 3 }), makeItem({ seat: 1 }), makeItem({ seat: 3 })];
    expect(occupiedSeats(items)).toEqual([1, 3]);
  });

  it('excludes items with no seat', () => {
    const items = [makeItem({ seat: 2 }), makeItem()];
    expect(occupiedSeats(items)).toEqual([2]);
  });
});

describe('groupBySeat', () => {
  it('groups items by seat number', () => {
    const items = [
      makeItem({ id: 'a', seat: 1 }),
      makeItem({ id: 'b', seat: 2 }),
      makeItem({ id: 'c', seat: 1 }),
    ];
    const grouped = groupBySeat(items);
    expect(grouped[1]).toHaveLength(2);
    expect(grouped[2]).toHaveLength(1);
  });

  it('places unassigned items under key 0', () => {
    const items = [makeItem({ id: 'a' }), makeItem({ id: 'b', seat: 1 })];
    const grouped = groupBySeat(items);
    expect(grouped[0]).toHaveLength(1);
    expect(grouped[0][0].id).toBe('a');
  });
});

describe('seatSubtotal', () => {
  it('sums price * quantity for items at a seat', () => {
    const items = [
      makeItem({ seat: 1, price: 10, quantity: 2 }),
      makeItem({ seat: 1, price: 5, quantity: 1 }),
      makeItem({ seat: 2, price: 8, quantity: 1 }),
    ];
    expect(seatSubtotal(items, 1)).toBeCloseTo(25);
  });

  it('returns 0 for a seat with no items', () => {
    expect(seatSubtotal([makeItem({ seat: 1, price: 10, quantity: 1 })], 2)).toBe(0);
  });
});

describe('unassignedSeatItems', () => {
  it('returns only items without a seat', () => {
    const items = [makeItem({ id: 'a', seat: 1 }), makeItem({ id: 'b' })];
    const result = unassignedSeatItems(items);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe('b');
  });
});
