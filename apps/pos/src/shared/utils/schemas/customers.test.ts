import { describe, it, expect } from 'vitest';
import { customerSchema } from './customers';

describe('customerSchema', () => {
  const base = { id: '42', name: 'Jane Doe' };

  it('parses a minimal valid customer', () => {
    const result = customerSchema.parse(base);
    expect(result.id).toBe('42');
    expect(result.name).toBe('Jane Doe');
  });

  it('coerces a numeric id to string', () => {
    const result = customerSchema.parse({ ...base, id: 99 });
    expect(result.id).toBe('99');
  });

  it('defaults name to "Customer" when missing', () => {
    const result = customerSchema.parse({ ...base, name: undefined });
    expect(result.name).toBe('Customer');
  });

  it('preserves phone when present', () => {
    const result = customerSchema.parse({ ...base, phone: '+1-555-0100' });
    expect(result.phone).toBe('+1-555-0100');
  });

  it('preserves email when present', () => {
    const result = customerSchema.parse({ ...base, email: 'jane@example.com' });
    expect(result.email).toBe('jane@example.com');
  });

  it('preserves loyaltyPoints when present', () => {
    const result = customerSchema.parse({ ...base, loyaltyPoints: 150 });
    expect(result.loyaltyPoints).toBe(150);
  });

  it('preserves lastVisit when present', () => {
    const iso = '2024-06-01T00:00:00Z';
    const result = customerSchema.parse({ ...base, lastVisit: iso });
    expect(result.lastVisit).toBe(iso);
  });

  it('keeps phone as null when explicitly null', () => {
    const result = customerSchema.parse({ ...base, phone: null });
    expect(result.phone).toBeNull();
  });

  it('keeps email as null when explicitly null', () => {
    const result = customerSchema.parse({ ...base, email: null });
    expect(result.email).toBeNull();
  });
});
