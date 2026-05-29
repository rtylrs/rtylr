import { describe, it, expect } from 'vitest';
import { recurringInvoiceSchema } from '../recurringInvoices';

describe('recurringInvoiceSchema', () => {
  it('coerces numeric id to string', () => {
    const result = recurringInvoiceSchema.parse({ id: 7, client: 'Acme' });
    expect(result.id).toBe('7');
  });

  it('defaults client to empty string when missing', () => {
    const result = recurringInvoiceSchema.parse({ id: '1' });
    expect(result.client).toBe('');
  });

  it('parses a full recurring invoice', () => {
    const result = recurringInvoiceSchema.parse({
      id: 'ri-10',
      client: 'Corp X',
      amount: 800,
      cadence: 'monthly',
      nextDate: '2026-07-01',
      status: 'Active',
    });
    expect(result.client).toBe('Corp X');
    expect(result.amount).toBe(800);
    expect(result.cadence).toBe('monthly');
  });

  it('allows null optional fields', () => {
    const result = recurringInvoiceSchema.parse({
      id: '2',
      client: 'Y',
      amount: null,
      cadence: null,
    });
    expect(result.amount).toBeNull();
    expect(result.cadence).toBeNull();
  });

  it('passes through unknown fields', () => {
    const result = recurringInvoiceSchema.parse({ id: '3', client: 'Z', currency: 'USD' });
    expect((result as Record<string, unknown>)['currency']).toBe('USD');
  });
});
