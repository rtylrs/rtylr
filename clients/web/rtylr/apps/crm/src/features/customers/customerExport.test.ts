import { describe, expect, it } from 'vitest';
import type { Customer } from '../../shared/schemas/customers';
import { customersToCsv } from './customerExport';

const thresholds = { Bronze: 0, Silver: 500, Gold: 1000, Platinum: 2500 } as const;

describe('customersToCsv', () => {
  it('escapes quoted values and protects formula-like customer fields', () => {
    const customer = {
      name: '=HYPERLINK("https://example.com")',
      email: '+malicious@example.com',
      phone: '555-1234',
      totalSpent: -12.5,
      visits: 3,
      loyaltyPoints: 10,
      loyaltyTier: 'bronze',
    } as Customer;

    expect(customersToCsv([customer], thresholds)).toBe(
      '"Name","Email","Phone","Spent","Visits","Points","Tier"\n' +
        '"\'=HYPERLINK(""https://example.com"")","\'+malicious@example.com","555-1234","-12.5","3","10","Bronze"'
    );
  });

  it('keeps ordinary customer values unchanged', () => {
    const customer = {
      name: 'Jane Doe',
      email: 'jane@example.com',
      phone: '555-1234',
      totalSpent: 120,
      visits: 4,
      loyaltyPoints: 500,
      loyaltyTier: 'silver',
    } as Customer;

    expect(customersToCsv([customer], thresholds)).toContain(
      '"Jane Doe","jane@example.com","555-1234","120","4","500","Silver"'
    );
  });
});
