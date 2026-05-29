// Customer analytics helpers for the insights customers feature.

import type { TopCustomer, LTVBucket, ChurnWindow } from '../hooks/queries/types';

/**
 * topCustomersBySpend returns customers sorted by lifetimeSpendMinorUnits descending.
 */
export function topCustomersBySpend(customers: TopCustomer[]): TopCustomer[] {
  return [...customers].sort((a, b) => b.lifetimeSpendMinorUnits - a.lifetimeSpendMinorUnits);
}

/**
 * avgOrderValue returns the average spend per order for a customer.
 * Returns 0 when orderCount is 0.
 */
export function avgOrderValue(customer: TopCustomer): number {
  if (customer.orderCount === 0) return 0;
  return customer.lifetimeSpendMinorUnits / customer.orderCount;
}

/**
 * totalCustomersInBuckets sums customerCount across all LTV buckets.
 */
export function totalCustomersInBuckets(buckets: LTVBucket[]): number {
  return buckets.reduce((sum, b) => sum + b.customerCount, 0);
}

/**
 * highValueBuckets returns LTV buckets where lowerMinorUnits meets or exceeds threshold.
 */
export function highValueBuckets(buckets: LTVBucket[], thresholdMinorUnits: number): LTVBucket[] {
  return buckets.filter((b) => b.lowerMinorUnits >= thresholdMinorUnits);
}

/**
 * lowestChurnWindow returns the ChurnWindow with the smallest churnPct.
 * Returns undefined for empty input.
 */
export function lowestChurnWindow(windows: ChurnWindow[]): ChurnWindow | undefined {
  if (windows.length === 0) return undefined;
  return windows.reduce((best, w) => (w.churnPct < best.churnPct ? w : best));
}

/**
 * churnRisk classifies a churn percentage as 'high', 'medium', or 'low'.
 */
export function churnRisk(churnPct: number): 'high' | 'medium' | 'low' {
  if (churnPct >= 20) return 'high';
  if (churnPct >= 10) return 'medium';
  return 'low';
}
