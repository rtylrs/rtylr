// Pure helpers for seat-based order management.

import type { OrderLineItem } from '../types/pos';

/** Return true if an item has been assigned to a seat. */
export function hasSeat(item: OrderLineItem): boolean {
  return typeof item.seat === 'number' && item.seat >= 1;
}

/** Return all items assigned to a specific seat number. */
export function itemsBySeat(items: OrderLineItem[], seat: number): OrderLineItem[] {
  return items.filter((i) => i.seat === seat);
}

/** Return the set of unique seat numbers present in the item list. */
export function occupiedSeats(items: OrderLineItem[]): number[] {
  const seats = new Set<number>();
  for (const item of items) {
    if (hasSeat(item)) seats.add(item.seat as number);
  }
  return Array.from(seats).sort((a, b) => a - b);
}

/** Group items by their seat number. Items without a seat go under key 0. */
export function groupBySeat(items: OrderLineItem[]): Record<number, OrderLineItem[]> {
  const result: Record<number, OrderLineItem[]> = {};
  for (const item of items) {
    const key = hasSeat(item) ? (item.seat as number) : 0;
    if (!result[key]) result[key] = [];
    result[key].push(item);
  }
  return result;
}

/** Return the subtotal for a specific seat. */
export function seatSubtotal(items: OrderLineItem[], seat: number): number {
  return itemsBySeat(items, seat).reduce(
    (sum, i) => sum + (Number.isFinite(i.price) ? i.price * i.quantity : 0),
    0,
  );
}

/** Return items that have not been assigned to any seat. */
export function unassignedSeatItems(items: OrderLineItem[]): OrderLineItem[] {
  return items.filter((i) => !hasSeat(i));
}
