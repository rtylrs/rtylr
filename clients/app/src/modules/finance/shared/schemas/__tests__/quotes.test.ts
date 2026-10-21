import { describe, it, expect } from 'vitest';
import { quoteSchema } from '../quotes';

describe('quoteSchema', () => {
  it('coerces numeric id to string', () => {
    const result = quoteSchema.parse({ id: 99, client: 'Beta' });
    expect(result.id).toBe('99');
  });

  it('defaults client to empty string when missing', () => {
    const result = quoteSchema.parse({ id: '1' });
    expect(result.client).toBe('');
  });

  it('accepts a full valid quote', () => {
    const result = quoteSchema.parse({
      id: 'q-3',
      client: 'Gamma Inc',
      amount: 1200,
      status: 'Accepted',
      expiryDate: '2026-08-01',
    });
    expect(result.id).toBe('q-3');
    expect(result.amount).toBe(1200);
    expect(result.status).toBe('Accepted');
  });

  it('allows null status and amount', () => {
    const result = quoteSchema.parse({ id: '4', client: 'Delta', amount: null, status: null });
    expect(result.amount).toBeNull();
    expect(result.status).toBeNull();
  });

  it('stores expiresAt when present', () => {
    const result = quoteSchema.parse({ id: '5', client: 'E', expiresAt: '2026-12-31' });
    expect(result.expiresAt).toBe('2026-12-31');
  });
});
