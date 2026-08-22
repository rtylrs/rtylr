import { describe, expect, it } from 'vitest';
import { daysAgo } from '../../shared/lib/dateUtils';
import { isAtRiskCustomer } from './Dashboard';

describe('isAtRiskCustomer', () => {
  it('treats a local date-only value as at risk after the threshold', () => {
    expect(isAtRiskCustomer(daysAgo(31), 30)).toBe(true);
    expect(isAtRiskCustomer(daysAgo(29), 30)).toBe(false);
  });

  it('handles relative and missing activity labels', () => {
    expect(isAtRiskCustomer('today')).toBe(false);
    expect(isAtRiskCustomer('45 days ago')).toBe(true);
    expect(isAtRiskCustomer('Never')).toBe(true);
    expect(isAtRiskCustomer('')).toBe(true);
    expect(isAtRiskCustomer('not-a-date')).toBe(true);
  });
});
