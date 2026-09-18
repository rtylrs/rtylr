import { describe, it, expect } from 'vitest';
import {
  displayName,
  hasLoyaltyPoints,
  loyaltyBalance,
  hasPhone,
  hasEmail,
  firstName,
  shortName,
  isReturningCustomer,
  sortByName,
} from './customerUtils';
import type { Customer } from '../types/pos';

const makeCustomer = (overrides: Partial<Customer> = {}): Customer => ({
  id: 'c1',
  name: 'Jane Doe',
  ...overrides,
});

describe('displayName', () => {
  it('returns the customer name when available', () => {
    expect(displayName(makeCustomer({ name: 'Jane Doe' }))).toBe('Jane Doe');
  });

  it('returns the fallback for null customer', () => {
    expect(displayName(null)).toBe('Guest');
  });

  it('returns the fallback for an empty name', () => {
    expect(displayName(makeCustomer({ name: '   ' }))).toBe('Guest');
  });

  it('accepts a custom fallback', () => {
    expect(displayName(null, 'Walk-in')).toBe('Walk-in');
  });
});

describe('hasLoyaltyPoints', () => {
  it('returns true when loyaltyPoints is a positive number', () => {
    expect(hasLoyaltyPoints(makeCustomer({ loyaltyPoints: 100 }))).toBe(true);
  });

  it('returns false when loyaltyPoints is 0', () => {
    expect(hasLoyaltyPoints(makeCustomer({ loyaltyPoints: 0 }))).toBe(false);
  });

  it('returns false when loyaltyPoints is undefined', () => {
    expect(hasLoyaltyPoints(makeCustomer())).toBe(false);
  });
});

describe('loyaltyBalance', () => {
  it('returns the loyalty points', () => {
    expect(loyaltyBalance(makeCustomer({ loyaltyPoints: 250 }))).toBe(250);
  });

  it('returns 0 when loyaltyPoints is undefined', () => {
    expect(loyaltyBalance(makeCustomer())).toBe(0);
  });

  it('floors at 0 for negative values', () => {
    expect(loyaltyBalance(makeCustomer({ loyaltyPoints: -10 }))).toBe(0);
  });
});

describe('hasPhone', () => {
  it('returns true when phone is a non-empty string', () => {
    expect(hasPhone(makeCustomer({ phone: '555-1234' }))).toBe(true);
  });

  it('returns false when phone is undefined', () => {
    expect(hasPhone(makeCustomer())).toBe(false);
  });

  it('returns false when phone is an empty string', () => {
    expect(hasPhone(makeCustomer({ phone: '' }))).toBe(false);
  });
});

describe('hasEmail', () => {
  it('returns true when email contains @', () => {
    expect(hasEmail(makeCustomer({ email: 'jane@example.com' }))).toBe(true);
  });

  it('returns false when email is undefined', () => {
    expect(hasEmail(makeCustomer())).toBe(false);
  });

  it('returns false when email lacks @', () => {
    expect(hasEmail(makeCustomer({ email: 'notanemail' }))).toBe(false);
  });
});

describe('firstName', () => {
  it('returns the first word of the customer name', () => {
    expect(firstName(makeCustomer({ name: 'Jane Doe' }))).toBe('Jane');
  });

  it('returns the full name when there is only one word', () => {
    expect(firstName(makeCustomer({ name: 'Jane' }))).toBe('Jane');
  });
});

describe('shortName', () => {
  it('returns first name and last initial', () => {
    expect(shortName(makeCustomer({ name: 'Jane Doe' }))).toBe('Jane D.');
  });

  it('returns just the name for a single-word name', () => {
    expect(shortName(makeCustomer({ name: 'Jane' }))).toBe('Jane');
  });
});

describe('isReturningCustomer', () => {
  it('returns true when lastVisit is set', () => {
    expect(isReturningCustomer(makeCustomer({ lastVisit: '2024-06-01' }))).toBe(true);
  });

  it('returns false when lastVisit is undefined', () => {
    expect(isReturningCustomer(makeCustomer())).toBe(false);
  });
});

describe('sortByName', () => {
  it('sorts customers alphabetically by name', () => {
    const customers = [
      makeCustomer({ id: 'b', name: 'Zoe' }),
      makeCustomer({ id: 'a', name: 'Alice' }),
    ];
    const sorted = sortByName(customers);
    expect(sorted[0].name).toBe('Alice');
    expect(sorted[1].name).toBe('Zoe');
  });

  it('does not mutate the original array', () => {
    const customers = [makeCustomer({ id: 'b', name: 'Zoe' }), makeCustomer({ id: 'a', name: 'Alice' })];
    sortByName(customers);
    expect(customers[0].name).toBe('Zoe');
  });
});
