import { describe, it, expect } from 'vitest';
import {
  subscriptionStatusLabel,
  totalMrr,
  annualizedRevenue,
  subscriptionsByStatus,
  churnedSubscriptions,
  trialConversionRate,
} from './subscription';
import type { Subscription } from './subscription';

function makeSub(
  status: 'trial' | 'active' | 'past_due' | 'cancelled' | 'paused',
  mrr = 99
): Subscription {
  return {
    id: 's1',
    customerId: 'c1',
    plan: 'pro',
    status,
    interval: 'monthly',
    mrr,
    startDate: '2026-01-01',
  };
}

const subs = [
  makeSub('active', 99),
  makeSub('active', 199),
  makeSub('cancelled', 49),
  makeSub('trial', 0),
  makeSub('past_due', 99),
];

describe('subscriptionStatusLabel', () => {
  it('returns display labels for known statuses', () => {
    expect(subscriptionStatusLabel('active')).toBe('Active');
    expect(subscriptionStatusLabel('past_due')).toBe('Past Due');
  });

  it('returns raw value for unknown status', () => {
    expect(subscriptionStatusLabel('suspended')).toBe('suspended');
  });
});

describe('totalMrr', () => {
  it('sums MRR of active subscriptions only', () => {
    expect(totalMrr(subs)).toBe(298);
  });

  it('returns 0 for empty array', () => {
    expect(totalMrr([])).toBe(0);
  });
});

describe('annualizedRevenue', () => {
  it('multiplies MRR by 12', () => {
    expect(annualizedRevenue(1000)).toBe(12000);
  });
});

describe('subscriptionsByStatus', () => {
  it('groups subscriptions by status', () => {
    const grouped = subscriptionsByStatus(subs);
    expect(grouped['active']).toHaveLength(2);
    expect(grouped['cancelled']).toHaveLength(1);
  });
});

describe('churnedSubscriptions', () => {
  it('returns cancelled subscriptions', () => {
    expect(churnedSubscriptions(subs)).toHaveLength(1);
  });
});

describe('trialConversionRate', () => {
  it('calculates conversion rate', () => {
    const rate = trialConversionRate(subs);
    expect(rate).toBeGreaterThan(0);
    expect(rate).toBeLessThanOrEqual(100);
  });

  it('returns 0 for empty array', () => {
    expect(trialConversionRate([])).toBe(0);
  });
});
