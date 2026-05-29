import { describe, it, expect } from 'vitest';
import {
  formatSalaryRange,
  isOfferExpired,
  offerStatusLabel,
  annualToMonthly,
} from './offerLetter';

describe('formatSalaryRange', () => {
  it('formats a salary range in USD', () => {
    const result = formatSalaryRange(50000, 70000, 'USD');
    expect(result).toContain('50,000');
    expect(result).toContain('70,000');
    expect(result).toContain('-');
  });
});

describe('isOfferExpired', () => {
  it('returns true for a past date', () => {
    expect(isOfferExpired('2020-01-01T00:00:00Z')).toBe(true);
  });

  it('returns false for a future date', () => {
    expect(isOfferExpired('2099-01-01T00:00:00Z')).toBe(false);
  });
});

describe('offerStatusLabel', () => {
  it('returns correct labels for known statuses', () => {
    expect(offerStatusLabel('pending')).toBe('Pending Signature');
    expect(offerStatusLabel('accepted')).toBe('Accepted');
    expect(offerStatusLabel('expired')).toBe('Expired');
  });

  it('returns status itself for unknown values', () => {
    expect(offerStatusLabel('custom')).toBe('custom');
  });
});

describe('annualToMonthly', () => {
  it('divides annual by 12 and rounds', () => {
    expect(annualToMonthly(120000)).toBe(10000);
    expect(annualToMonthly(50000)).toBe(4167);
  });
});
