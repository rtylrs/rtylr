import { describe, it, expect } from 'vitest';
import { billSchema } from '../bills';

describe('billSchema', () => {
  it('coerces numeric id to string', () => {
    const result = billSchema.parse({ id: 42, supplier: 'ACME' });
    expect(result.id).toBe('42');
  });

  it('defaults supplier to empty string when missing', () => {
    const result = billSchema.parse({ id: '1' });
    expect(result.supplier).toBe('');
  });

  it('accepts a full valid bill', () => {
    const result = billSchema.parse({
      id: 'bill-7',
      supplier: 'Supplier Co',
      amount: 500,
      status: 'Unpaid',
      dueDate: '2026-07-01',
    });
    expect(result.id).toBe('bill-7');
    expect(result.supplier).toBe('Supplier Co');
    expect(result.amount).toBe(500);
  });

  it('allows nullable optional fields', () => {
    const result = billSchema.parse({ id: '2', supplier: 'X', amount: null, status: null });
    expect(result.amount).toBeNull();
    expect(result.status).toBeNull();
  });

  it('passes through unknown fields', () => {
    const result = billSchema.parse({ id: '3', supplier: 'Y', extra: 'data' });
    expect((result as Record<string, unknown>)['extra']).toBe('data');
  });
});
