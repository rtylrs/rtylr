// Financial analytics helpers for the insights financial feature.

import type { CashFlowPoint, ExpenseByCategory, InvoiceAgingBucket } from '../hooks/queries/types';

/**
 * netCashFlow returns the total net cash flow across the series.
 */
export function netCashFlow(series: CashFlowPoint[]): number {
  return series.reduce((sum, p) => sum + p.inflowMinorUnits - p.outflowMinorUnits, 0);
}

/**
 * peakInflowDay returns the CashFlowPoint with the highest inflowMinorUnits.
 * Returns undefined for an empty series.
 */
export function peakInflowDay(series: CashFlowPoint[]): CashFlowPoint | undefined {
  if (series.length === 0) return undefined;
  return series.reduce((best, p) => (p.inflowMinorUnits > best.inflowMinorUnits ? p : best));
}

/**
 * topExpenseCategories returns expense categories sorted by amountMinorUnits descending.
 */
export function topExpenseCategories(
  categories: ExpenseByCategory[],
  n: number
): ExpenseByCategory[] {
  return [...categories].sort((a, b) => b.amountMinorUnits - a.amountMinorUnits).slice(0, n);
}

/**
 * totalExpenses returns the sum of amountMinorUnits across all expense categories.
 */
export function totalExpenses(categories: ExpenseByCategory[]): number {
  return categories.reduce((sum, c) => sum + c.amountMinorUnits, 0);
}

/**
 * overdueInvoiceTotal returns the sum of amountMinorUnits for buckets whose
 * label is not '0-30 days'.
 */
export function overdueInvoiceTotal(buckets: InvoiceAgingBucket[]): number {
  return buckets
    .filter((b) => b.bucket !== '0-30 days')
    .reduce((sum, b) => sum + b.amountMinorUnits, 0);
}

/**
 * largestAgingBucket returns the bucket with the highest amountMinorUnits.
 * Returns undefined for an empty array.
 */
export function largestAgingBucket(buckets: InvoiceAgingBucket[]): InvoiceAgingBucket | undefined {
  if (buckets.length === 0) return undefined;
  return buckets.reduce((best, b) => (b.amountMinorUnits > best.amountMinorUnits ? b : best));
}
