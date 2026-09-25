import { describe, it, expect } from 'vitest';
import { recurringInvoiceSchema } from '../recurringInvoices';

describe('recurringInvoiceSchema', () => {
  it('parses the onekit payload shape', () => {
    const result = recurringInvoiceSchema.parse({
      id: '10',
      templateInvoiceId: '42',
      frequency: 'monthly',
      nextInvoiceDate: '2026-07-01',
      isActive: true,
    });
    expect(result.id).toBe('10');
    expect(result.templateInvoiceId).toBe('42');
    expect(result.frequency).toBe('monthly');
    expect(result.isActive).toBe(true);
  });

  it('allows null/optional date fields', () => {
    const result = recurringInvoiceSchema.parse({
      id: '2',
      frequency: 'weekly',
      nextInvoiceDate: null,
      endDate: null,
      isActive: false,
    });
    expect(result.nextInvoiceDate).toBeNull();
    expect(result.endDate).toBeNull();
  });

  it('passes through unknown fields', () => {
    const result = recurringInvoiceSchema.parse({
      id: '3',
      frequency: 'yearly',
      isActive: true,
      ownerId: '99',
      somethingExtra: 'kept',
    });
    expect((result as Record<string, unknown>)['somethingExtra']).toBe('kept');
  });
});
