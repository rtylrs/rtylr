import { describe, it, expect } from 'vitest';
import {
  getOutstandingAmount,
  isFullyPaid,
  isOverdue,
  computeInvoiceTotals,
  getStatusLabel,
  getStatusColor,
} from './invoiceUtils';

describe('getOutstandingAmount', () => {
  it('returns full amount when nothing is paid', () => {
    expect(getOutstandingAmount({ amount: 1000 })).toBe(1000);
  });

  it('returns difference between amount and paid', () => {
    expect(getOutstandingAmount({ amount: 1000, paid: 400 })).toBe(600);
  });

  it('returns 0 when fully paid', () => {
    expect(getOutstandingAmount({ amount: 500, paid: 500 })).toBe(0);
  });

  it('returns 0 when overpaid (clamps to 0)', () => {
    expect(getOutstandingAmount({ amount: 500, paid: 600 })).toBe(0);
  });

  it('treats null paid as 0', () => {
    expect(getOutstandingAmount({ amount: 200, paid: null })).toBe(200);
  });
});

describe('isFullyPaid', () => {
  it('returns true when outstanding is 0', () => {
    expect(isFullyPaid({ amount: 300, paid: 300 })).toBe(true);
  });

  it('returns false when outstanding > 0', () => {
    expect(isFullyPaid({ amount: 300, paid: 100 })).toBe(false);
  });

  it('returns false when nothing is paid', () => {
    expect(isFullyPaid({ amount: 100 })).toBe(false);
  });
});

describe('isOverdue', () => {
  it('returns true when due date is before today', () => {
    expect(isOverdue({ amount: 100, dueDate: '2020-01-01' }, '2026-06-01')).toBe(true);
  });

  it('returns false when due date is today or future', () => {
    expect(isOverdue({ amount: 100, dueDate: '2026-06-01' }, '2026-06-01')).toBe(false);
    expect(isOverdue({ amount: 100, dueDate: '2026-07-01' }, '2026-06-01')).toBe(false);
  });

  it('returns false when status is Paid', () => {
    expect(isOverdue({ amount: 100, dueDate: '2020-01-01', status: 'Paid' }, '2026-06-01')).toBe(
      false
    );
  });

  it('returns false when no dueDate', () => {
    expect(isOverdue({ amount: 100 }, '2026-06-01')).toBe(false);
  });
});

describe('computeInvoiceTotals', () => {
  it('computes subtotal from line items with amount', () => {
    const items = [{ amount: 500 }, { amount: 300 }];
    const totals = computeInvoiceTotals(items);
    expect(totals.subtotal).toBe(800);
  });

  it('computes subtotal from qty * unitPrice', () => {
    const items = [{ quantity: 3, unitPrice: 100 }];
    const totals = computeInvoiceTotals(items);
    expect(totals.subtotal).toBe(300);
  });

  it('applies discount before tax', () => {
    const items = [{ amount: 1000 }];
    const totals = computeInvoiceTotals(items, 100, 10);
    expect(totals.discount).toBe(100);
    expect(totals.tax).toBeCloseTo(90);
    expect(totals.total).toBeCloseTo(990);
  });

  it('clamps discount to subtotal', () => {
    const items = [{ amount: 200 }];
    const totals = computeInvoiceTotals(items, 500);
    expect(totals.discount).toBe(200);
    expect(totals.total).toBe(0);
  });

  it('returns zero tax when taxRate is 0', () => {
    const items = [{ amount: 1000 }];
    const totals = computeInvoiceTotals(items, 0, 0);
    expect(totals.tax).toBe(0);
    expect(totals.total).toBe(1000);
  });
});

describe('getStatusLabel', () => {
  it('returns human readable label for each status', () => {
    expect(getStatusLabel('Draft')).toBe('Draft');
    expect(getStatusLabel('Sent')).toBe('Sent');
    expect(getStatusLabel('Paid')).toBe('Paid');
    expect(getStatusLabel('Overdue')).toBe('Overdue');
  });
});

describe('getStatusColor', () => {
  it('returns correct color for each status', () => {
    expect(getStatusColor('Draft')).toBe('gray');
    expect(getStatusColor('Sent')).toBe('blue');
    expect(getStatusColor('Paid')).toBe('green');
    expect(getStatusColor('Overdue')).toBe('red');
  });
});
