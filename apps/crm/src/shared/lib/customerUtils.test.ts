import { describe, it, expect } from 'vitest';
import {
  hasEmailConsent,
  hasSmsConsent,
  daysSinceLastVisit,
  isAtRisk,
  getTags,
  hasTag,
  customerLabel,
  formatCustomerDate,
} from './customerUtils';
import { customerSchema, type Customer } from '../schemas/customers';

const makeCustomer = (overrides: Record<string, unknown> = {}): Customer =>
  customerSchema.parse({
    id: 1,
    firstName: 'Jane',
    lastName: 'Doe',
    email: 'jane@example.com',
    phone: '+1 555-0100',
    totalSpent: 500,
    visitCount: 10,
    loyaltyPoints: 200,
    lastVisitAt: '2024-01-01',
    ...overrides,
  });

describe('hasEmailConsent', () => {
  it('returns true when email consent is true', () => {
    const c = makeCustomer({ consent: { email: true, sms: false, marketing: false } });
    expect(hasEmailConsent(c)).toBe(true);
  });

  it('returns false when email consent is false', () => {
    const c = makeCustomer({ consent: { email: false, sms: false, marketing: false } });
    expect(hasEmailConsent(c)).toBe(false);
  });

  it('returns false when consent is not set', () => {
    const c = makeCustomer();
    expect(hasEmailConsent(c)).toBe(false);
  });
});

describe('hasSmsConsent', () => {
  it('returns true when sms consent is true', () => {
    const c = makeCustomer({ consent: { email: false, sms: true, marketing: false } });
    expect(hasSmsConsent(c)).toBe(true);
  });

  it('returns false when sms consent is missing', () => {
    expect(hasSmsConsent(makeCustomer())).toBe(false);
  });
});

describe('daysSinceLastVisit', () => {
  it('returns 0 for a visit today', () => {
    const today = new Date();
    const dateStr = today.toISOString().split('T')[0];
    const c = makeCustomer({ lastVisitAt: dateStr });
    expect(daysSinceLastVisit(c, today)).toBe(0);
  });

  it('returns the correct number of days', () => {
    const now = new Date('2024-05-10');
    const c = makeCustomer({ lastVisitAt: '2024-04-30' });
    expect(daysSinceLastVisit(c, now)).toBe(10);
  });

  it('treats a missing or malformed visit date as infinitely old', () => {
    expect(daysSinceLastVisit(makeCustomer({ lastVisitAt: '' }), new Date('2024-05-10'))).toBe(
      Infinity
    );
    expect(
      daysSinceLastVisit(makeCustomer({ lastVisitAt: 'not-a-date' }), new Date('2024-05-10'))
    ).toBe(Infinity);
  });
});

describe('isAtRisk', () => {
  it('returns true when days since last visit >= atRiskDays', () => {
    const now = new Date('2024-05-10');
    const c = makeCustomer({ lastVisitAt: '2024-04-10' });
    expect(isAtRisk(c, 30, now)).toBe(true);
  });

  it('returns false when days since last visit < atRiskDays', () => {
    const now = new Date('2024-05-10');
    const c = makeCustomer({ lastVisitAt: '2024-04-25' });
    expect(isAtRisk(c, 30, now)).toBe(false);
  });
});

describe('getTags', () => {
  it('returns the customer tags', () => {
    const c = makeCustomer({ tags: ['vip', 'regular'] });
    expect(getTags(c)).toEqual(['vip', 'regular']);
  });

  it('returns empty array when tags are not set', () => {
    expect(getTags(makeCustomer())).toEqual([]);
  });
});

describe('hasTag', () => {
  it('returns true when customer has the tag', () => {
    const c = makeCustomer({ tags: ['vip'] });
    expect(hasTag(c, 'vip')).toBe(true);
  });

  it('returns false when customer does not have the tag', () => {
    const c = makeCustomer({ tags: ['regular'] });
    expect(hasTag(c, 'vip')).toBe(false);
  });
});

describe('customerLabel', () => {
  it('formats name and email together', () => {
    const c = makeCustomer({ firstName: 'Jane', lastName: 'Doe', email: 'jane@example.com' });
    expect(customerLabel(c)).toBe('Jane Doe <jane@example.com>');
  });
});

describe('formatCustomerDate', () => {
  it('formats ISO timestamps for display', () => {
    expect(formatCustomerDate('2024-01-15')).toBe('Jan 15, 2024');
  });

  it('preserves supported relative labels', () => {
    expect(formatCustomerDate('45 days ago')).toBe('45 days ago');
  });

  it('uses clear fallbacks for missing and malformed dates', () => {
    expect(formatCustomerDate('')).toBe('Never');
    expect(formatCustomerDate('not-a-date')).toBe('Unknown date');
  });
});
