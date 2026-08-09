import { describe, it, expect } from 'vitest';
import {
  expiryStatus,
  isExpired,
  needsAttention,
  sortByExpiry,
  EXPIRY_WARNING_DAYS,
  EXPIRY_CRITICAL_DAYS,
} from './expiryUtils';

const TODAY = new Date('2026-06-01T00:00:00.000Z');

describe('expiryStatus', () => {
  it('returns "expired" for a past date', () => {
    expect(expiryStatus('2026-05-01', TODAY)).toBe('expired');
  });

  it('returns "critical" when expiry is within the critical window', () => {
    const critical = new Date(TODAY);
    critical.setDate(critical.getDate() + EXPIRY_CRITICAL_DAYS - 1);
    expect(expiryStatus(critical.toISOString(), TODAY)).toBe('critical');
  });

  it('returns "warning" when expiry is between critical and warning window', () => {
    const warning = new Date(TODAY);
    warning.setDate(warning.getDate() + EXPIRY_WARNING_DAYS - 1);
    expect(expiryStatus(warning.toISOString(), TODAY)).toBe('warning');
  });

  it('returns "ok" for a date well in the future', () => {
    expect(expiryStatus('2027-06-01', TODAY)).toBe('ok');
  });

  it('returns "expired" for today when time component makes it past', () => {
    const yesterday = new Date(TODAY);
    yesterday.setDate(yesterday.getDate() - 1);
    expect(expiryStatus(yesterday.toISOString(), TODAY)).toBe('expired');
  });
});

describe('isExpired', () => {
  it('returns true for a past date', () => {
    expect(isExpired('2025-01-01', TODAY)).toBe(true);
  });

  it('returns false for a future date', () => {
    expect(isExpired('2027-01-01', TODAY)).toBe(false);
  });
});

describe('needsAttention', () => {
  it('returns true for a critically close date', () => {
    const soon = new Date(TODAY);
    soon.setDate(soon.getDate() + 3);
    expect(needsAttention(soon.toISOString(), TODAY)).toBe(true);
  });

  it('returns true for a warning-range date', () => {
    const warning = new Date(TODAY);
    warning.setDate(warning.getDate() + 15);
    expect(needsAttention(warning.toISOString(), TODAY)).toBe(true);
  });

  it('returns false for a distant future date', () => {
    expect(needsAttention('2028-01-01', TODAY)).toBe(false);
  });

  it('returns false for an expired item', () => {
    expect(needsAttention('2025-01-01', TODAY)).toBe(false);
  });
});

describe('sortByExpiry', () => {
  const items = [
    { sku: 'C', expiryDate: '2026-09-01' },
    { sku: 'A', expiryDate: '2026-07-01' },
    { sku: 'B', expiryDate: '2026-08-01' },
  ];

  it('sorts items from earliest to latest expiry', () => {
    const sorted = sortByExpiry(items);
    expect(sorted.map((i) => i.sku)).toEqual(['A', 'B', 'C']);
  });

  it('does not mutate the original array', () => {
    const original = [...items];
    sortByExpiry(items);
    expect(items).toEqual(original);
  });

  it('pushes items with invalid dates to the end', () => {
    const withInvalid = [
      { sku: 'Bad', expiryDate: 'not-a-date' },
      { sku: 'Good', expiryDate: '2026-07-01' },
    ];
    const sorted = sortByExpiry(withInvalid);
    expect(sorted[0].sku).toBe('Good');
    expect(sorted[1].sku).toBe('Bad');
  });
});
