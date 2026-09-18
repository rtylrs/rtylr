// Pure helpers for order-level calculations.

import type { OrderLineItem } from '../types/pos';

/** Sum of all line item prices (price * quantity). */
export function orderSubtotal(items: OrderLineItem[]): number {
  return items.reduce((sum, item) => sum + item.price * item.quantity, 0);
}

/** Round a currency amount to two decimal places. */
export function roundCents(amount: number): number {
  return Math.round(amount * 100) / 100;
}

/** Calculate tax amount from a subtotal and a tax rate (0-1). */
export function calcTax(subtotal: number, taxRate: number): number {
  if (!Number.isFinite(subtotal) || subtotal < 0) return 0;
  if (!Number.isFinite(taxRate) || taxRate < 0) return 0;
  return roundCents(subtotal * taxRate);
}

/** Calculate the change due when tendering cash. Returns 0 if tendered < due. */
export function calcChange(tendered: number, amountDue: number): number {
  if (!Number.isFinite(tendered) || tendered < 0) return 0;
  if (!Number.isFinite(amountDue) || amountDue < 0) return 0;
  return Math.max(0, roundCents(tendered - amountDue));
}

/** Check whether a cash tender covers the full amount due. */
export function canCharge(tendered: number, amountDue: number): boolean {
  if (!Number.isFinite(tendered) || !Number.isFinite(amountDue)) return false;
  if (amountDue <= 0) return false;
  return tendered >= amountDue;
}

/** Apply a percentage discount (0-100) to a subtotal. */
export function applyPercentDiscount(subtotal: number, pct: number): number {
  if (!Number.isFinite(subtotal) || subtotal < 0) return 0;
  const clampedPct = Math.min(100, Math.max(0, pct));
  return roundCents(subtotal * (1 - clampedPct / 100));
}

/** Apply a flat discount to a subtotal, flooring at 0. */
export function applyFlatDiscount(subtotal: number, flatAmount: number): number {
  if (!Number.isFinite(subtotal) || subtotal < 0) return 0;
  if (!Number.isFinite(flatAmount) || flatAmount < 0) return subtotal;
  return Math.max(0, roundCents(subtotal - flatAmount));
}

/** Split a total evenly across N seats, distributing any rounding remainder
 *  to the first seat. */
export function splitEvenly(total: number, seats: number): number[] {
  if (!Number.isFinite(total) || total < 0) return [];
  const n = Math.floor(seats);
  if (n <= 1) return [roundCents(total)];
  const perSeat = Math.floor((total * 100) / n) / 100;
  const remainder = roundCents(total - perSeat * n);
  return [roundCents(perSeat + remainder), ...Array(n - 1).fill(perSeat)];
}
