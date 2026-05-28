import { describe, it, expect } from 'vitest';
import {
  customerAnalyticsSchema,
  topCustomerSchema,
  customerChurnSchema,
  customerLtvSchema,
} from './customers';

describe('customerAnalyticsSchema', () => {
  it('parses a full customer analytics object', () => {
    const data = {
      totalCustomers: 500,
      newCustomers: 30,
      returningCustomers: 470,
      churnRate: 0.05,
      avgLifetimeValue: 1200,
    };
    const result = customerAnalyticsSchema.parse(data);
    expect(result.totalCustomers).toBe(500);
    expect(result.churnRate).toBe(0.05);
  });

  it('accepts an empty object (all fields optional)', () => {
    expect(() => customerAnalyticsSchema.parse({})).not.toThrow();
  });
});

describe('topCustomerSchema', () => {
  it('coerces numeric id to string', () => {
    const result = topCustomerSchema.parse({ id: 42, name: 'Acme Corp' });
    expect(result.id).toBe('42');
  });

  it('keeps string id as string', () => {
    const result = topCustomerSchema.parse({ id: 'abc', name: 'Test' });
    expect(result.id).toBe('abc');
  });

  it('accepts nullable email and lastOrderAt', () => {
    const result = topCustomerSchema.parse({ id: 1, email: null, lastOrderAt: null });
    expect(result.email).toBeNull();
    expect(result.lastOrderAt).toBeNull();
  });

  it('accepts an object without id', () => {
    expect(() => topCustomerSchema.parse({ name: 'Anonymous' })).not.toThrow();
  });
});

describe('customerChurnSchema', () => {
  it('parses churn data correctly', () => {
    const data = { rate: 0.08, churnRate: 0.08, churned: 40, atRisk: 15, retained: 460 };
    const result = customerChurnSchema.parse(data);
    expect(result.churnRate).toBe(0.08);
    expect(result.churned).toBe(40);
  });

  it('accepts an empty object', () => {
    expect(() => customerChurnSchema.parse({})).not.toThrow();
  });
});

describe('customerLtvSchema', () => {
  it('parses LTV data correctly', () => {
    const data = { average: 1000, avgLtv: 1000, medianLtv: 800, topDecileLtv: 5000 };
    const result = customerLtvSchema.parse(data);
    expect(result.avgLtv).toBe(1000);
    expect(result.topDecileLtv).toBe(5000);
  });

  it('accepts an empty object', () => {
    expect(() => customerLtvSchema.parse({})).not.toThrow();
  });
});
