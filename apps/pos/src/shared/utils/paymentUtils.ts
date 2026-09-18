// Pure helpers for multi-payment and split-payment scenarios.

import type { PaymentEntry, PaymentMethod } from '../types/pos';

/** Sum the total amount across a set of payment entries. */
export function totalPaid(payments: PaymentEntry[]): number {
  return payments.reduce(
    (sum, p) => sum + (Number.isFinite(p.amount) ? p.amount : 0),
    0,
  );
}

/** Sum the total tips across a set of payment entries. */
export function totalTips(payments: PaymentEntry[]): number {
  return payments.reduce(
    (sum, p) => sum + (Number.isFinite(p.tipAmount) ? (p.tipAmount ?? 0) : 0),
    0,
  );
}

/** Return the remaining amount due given total paid and the order total. */
export function remainingDue(orderTotal: number, paid: number): number {
  if (!Number.isFinite(orderTotal) || !Number.isFinite(paid)) return 0;
  return Math.max(0, orderTotal - paid);
}

/** Return true if the order is fully paid. */
export function isFullyPaid(orderTotal: number, paid: number, tolerance = 0.001): boolean {
  if (!Number.isFinite(orderTotal) || !Number.isFinite(paid)) return false;
  return paid >= orderTotal - tolerance;
}

/** Group payment entries by method and sum amounts per method. */
export function paymentsByMethod(payments: PaymentEntry[]): Record<PaymentMethod, number> {
  const result: Record<PaymentMethod, number> = {
    cash: 0,
    card: 0,
    gift_card: 0,
    other: 0,
  };
  for (const p of payments) {
    if (Number.isFinite(p.amount)) {
      result[p.method] = (result[p.method] ?? 0) + p.amount;
    }
  }
  return result;
}

/** Return the count of payments by method. */
export function paymentMethodCount(payments: PaymentEntry[]): Record<PaymentMethod, number> {
  const result: Record<PaymentMethod, number> = {
    cash: 0,
    card: 0,
    gift_card: 0,
    other: 0,
  };
  for (const p of payments) {
    result[p.method] = (result[p.method] ?? 0) + 1;
  }
  return result;
}

/** Return true if any payment used cash. */
export function hasCashPayment(payments: PaymentEntry[]): boolean {
  return payments.some((p) => p.method === 'cash');
}

/** Return true if any payment used a gift card. */
export function hasGiftCardPayment(payments: PaymentEntry[]): boolean {
  return payments.some((p) => p.method === 'gift_card');
}

/** Return the largest single payment amount. */
export function largestPayment(payments: PaymentEntry[]): number {
  if (!payments.length) return 0;
  return Math.max(...payments.map((p) => (Number.isFinite(p.amount) ? p.amount : 0)));
}
