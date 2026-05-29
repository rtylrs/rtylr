// Subscription analytics helpers for the insights analytics feature.

export type SubscriptionStatus = 'trial' | 'active' | 'past_due' | 'cancelled' | 'paused';
export type BillingInterval = 'monthly' | 'annual' | 'quarterly';

export interface Subscription {
  id: string;
  customerId: string;
  plan: string;
  status: SubscriptionStatus;
  interval: BillingInterval;
  mrr: number;
  startDate: string;
  endDate?: string;
}

/**
 * subscriptionStatusLabel returns a display label for a subscription status.
 */
export function subscriptionStatusLabel(status: SubscriptionStatus | string): string {
  const labels: Record<string, string> = {
    trial: 'Trial',
    active: 'Active',
    past_due: 'Past Due',
    cancelled: 'Cancelled',
    paused: 'Paused',
  };
  return labels[status] ?? status;
}

/**
 * totalMrr returns the total monthly recurring revenue across active subscriptions.
 */
export function totalMrr(subscriptions: Subscription[]): number {
  return subscriptions.filter((s) => s.status === 'active').reduce((sum, s) => sum + s.mrr, 0);
}

/**
 * annualizedRevenue returns the annualized revenue from MRR.
 */
export function annualizedRevenue(mrr: number): number {
  return mrr * 12;
}

/**
 * subscriptionsByStatus groups subscriptions by their status.
 */
export function subscriptionsByStatus(
  subscriptions: Subscription[]
): Record<string, Subscription[]> {
  return subscriptions.reduce<Record<string, Subscription[]>>((acc, s) => {
    acc[s.status] = [...(acc[s.status] ?? []), s];
    return acc;
  }, {});
}

/**
 * churnedSubscriptions returns subscriptions with cancelled status.
 */
export function churnedSubscriptions(subscriptions: Subscription[]): Subscription[] {
  return subscriptions.filter((s) => s.status === 'cancelled');
}

/**
 * trialConversionRate returns the percentage of trials that converted to active.
 */
export function trialConversionRate(subscriptions: Subscription[]): number {
  const trials = subscriptions.filter(
    (s) => s.status === 'active' || s.status === 'cancelled' || s.status === 'past_due'
  );
  if (trials.length === 0) return 0;
  const converted = subscriptions.filter((s) => s.status === 'active').length;
  return Math.round((converted / trials.length) * 1000) / 10;
}
