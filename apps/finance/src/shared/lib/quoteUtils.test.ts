import { describe, it, expect } from 'vitest';
import {
  isAccepted,
  isTerminal,
  isExpired,
  daysUntilExpiry,
  getQuoteStatusLabel,
  resolveExpiryDate,
  quoteConversionRate,
  getQuoteStatusColor,
} from './quoteUtils';

describe('isAccepted', () => {
  it('returns true for Accepted status', () => {
    expect(isAccepted('Accepted')).toBe(true);
  });

  it('returns false for other statuses', () => {
    expect(isAccepted('Draft')).toBe(false);
    expect(isAccepted('Sent')).toBe(false);
    expect(isAccepted('Rejected')).toBe(false);
  });
});

describe('isTerminal', () => {
  it('returns true for Accepted, Rejected, Expired', () => {
    expect(isTerminal('Accepted')).toBe(true);
    expect(isTerminal('Rejected')).toBe(true);
    expect(isTerminal('Expired')).toBe(true);
  });

  it('returns false for Draft and Sent', () => {
    expect(isTerminal('Draft')).toBe(false);
    expect(isTerminal('Sent')).toBe(false);
  });
});

describe('isExpired', () => {
  it('returns true when expiry date is in the past', () => {
    expect(isExpired('2020-01-01', '2026-06-01')).toBe(true);
  });

  it('returns false when expiry date is today or future', () => {
    expect(isExpired('2026-06-01', '2026-06-01')).toBe(false);
    expect(isExpired('2026-12-31', '2026-06-01')).toBe(false);
  });

  it('returns false when expiry date is null or undefined', () => {
    expect(isExpired(null, '2026-06-01')).toBe(false);
    expect(isExpired(undefined, '2026-06-01')).toBe(false);
  });
});

describe('daysUntilExpiry', () => {
  it('returns positive number for future expiry', () => {
    expect(daysUntilExpiry('2026-06-20', '2026-06-15')).toBe(5);
  });

  it('returns 0 when expiry is today', () => {
    expect(daysUntilExpiry('2026-06-15', '2026-06-15')).toBe(0);
  });

  it('returns negative for past expiry', () => {
    expect(daysUntilExpiry('2026-06-10', '2026-06-15')).toBe(-5);
  });

  it('returns null when expiry date is not set', () => {
    expect(daysUntilExpiry(null)).toBeNull();
    expect(daysUntilExpiry(undefined)).toBeNull();
  });
});

describe('getQuoteStatusLabel', () => {
  it('returns correct label for each status', () => {
    expect(getQuoteStatusLabel('Draft')).toBe('Draft');
    expect(getQuoteStatusLabel('Sent')).toBe('Sent');
    expect(getQuoteStatusLabel('Accepted')).toBe('Accepted');
    expect(getQuoteStatusLabel('Rejected')).toBe('Rejected');
    expect(getQuoteStatusLabel('Expired')).toBe('Expired');
  });
});

describe('resolveExpiryDate', () => {
  it('returns expiryDate when available', () => {
    expect(resolveExpiryDate({ expiryDate: '2026-08-01', expiresAt: '2026-07-01' })).toBe(
      '2026-08-01'
    );
  });

  it('falls back to expiresAt when expiryDate is null', () => {
    expect(resolveExpiryDate({ expiryDate: null, expiresAt: '2026-07-01' })).toBe('2026-07-01');
  });

  it('returns null when both are null', () => {
    expect(resolveExpiryDate({ expiryDate: null, expiresAt: null })).toBeNull();
  });

  it('returns null when neither is set', () => {
    expect(resolveExpiryDate({})).toBeNull();
  });
});

describe('quoteConversionRate', () => {
  it('returns percentage of accepted quotes', () => {
    expect(quoteConversionRate(10, 3)).toBe(30);
    expect(quoteConversionRate(4, 1)).toBe(25);
  });

  it('returns 0 when total is 0', () => {
    expect(quoteConversionRate(0, 0)).toBe(0);
  });

  it('returns 100 when all accepted', () => {
    expect(quoteConversionRate(5, 5)).toBe(100);
  });
});

describe('getQuoteStatusColor', () => {
  it('returns correct color for each status', () => {
    expect(getQuoteStatusColor('Draft')).toBe('gray');
    expect(getQuoteStatusColor('Sent')).toBe('blue');
    expect(getQuoteStatusColor('Accepted')).toBe('green');
    expect(getQuoteStatusColor('Rejected')).toBe('red');
    expect(getQuoteStatusColor('Expired')).toBe('orange');
  });
});
