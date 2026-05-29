/**
 * Customer retention analysis helpers for CRM.
 */

export interface RetentionMetrics {
  total: number;
  retained: number;
  churned: number;
  retentionRate: number;
  churnRate: number;
}

/** Calculates retention metrics from total and retained counts. */
export function retentionMetrics(total: number, retained: number): RetentionMetrics {
  const safe = Math.max(0, Math.min(retained, total));
  const churned = total - safe;
  const retentionRate = total === 0 ? 0 : safe / total;
  const churnRate = total === 0 ? 0 : churned / total;
  return { total, retained: safe, churned, retentionRate, churnRate };
}

/** Returns a risk label based on days since last visit. */
export function riskLabel(daysSinceVisit: number): 'healthy' | 'at-risk' | 'churned' {
  if (daysSinceVisit <= 30) return 'healthy';
  if (daysSinceVisit <= 90) return 'at-risk';
  return 'churned';
}

/** Returns the customer lifetime value estimate. */
export function clv(
  averageOrderValue: number,
  purchaseFrequencyPerYear: number,
  averageLifespanYears: number
): number {
  return averageOrderValue * purchaseFrequencyPerYear * averageLifespanYears;
}

/** Returns the repeat purchase rate as a fraction. */
export function repeatPurchaseRate(totalCustomers: number, repeatCustomers: number): number {
  if (totalCustomers === 0) return 0;
  return Math.min(1, repeatCustomers / totalCustomers);
}
