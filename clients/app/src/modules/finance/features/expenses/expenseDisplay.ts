import type { CSSProperties } from 'react';

/** The values finance_expenses.payment_method accepts, in the order offered. */
export const PAYMENT_METHODS = ['card', 'cash', 'bank_transfer', 'other'] as const;

/** Badge colours for the expense review states. */
export const expenseStatusStyle: Record<string, CSSProperties> = {
  pending: { background: 'var(--finance-warning-soft)', color: 'var(--finance-warning)' },
  approved: { background: 'var(--finance-success-soft)', color: 'var(--finance-success)' },
  rejected: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
};

const LAST_PAYMENT_METHOD_KEY = 'finance.expenses.lastPaymentMethod';

/** The method used for the previous expense, so a card-only business picks it once. */
export function lastPaymentMethod(): string {
  try {
    const saved = window.localStorage.getItem(LAST_PAYMENT_METHOD_KEY);
    if (saved && (PAYMENT_METHODS as readonly string[]).includes(saved)) return saved;
  } catch {
    // Storage can be unavailable (private mode); fall through to the default.
  }
  return 'card';
}

export function rememberPaymentMethod(method: string): void {
  try {
    window.localStorage.setItem(LAST_PAYMENT_METHOD_KEY, method);
  } catch {
    // Not remembering is harmless.
  }
}
