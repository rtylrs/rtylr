import { describe, expect, it } from 'vitest';
import { formatPeriodRange, trialDaysLeft } from './useCrmSubscription';

describe('subscription date helpers', () => {
  it('returns zero for a missing or malformed trial end date', () => {
    expect(trialDaysLeft(null)).toBe(0);
    expect(trialDaysLeft('not-a-date')).toBe(0);
  });

  it('does not render malformed billing periods as undefined dates', () => {
    expect(formatPeriodRange('not-a-date', null)).toBe('');
    expect(formatPeriodRange('2026-08-01', 'not-a-date')).toBe('');
  });
});
