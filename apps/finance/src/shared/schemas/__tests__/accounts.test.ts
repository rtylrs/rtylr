import { describe, it, expect } from 'vitest';
import { bankAccountSchema, chartOfAccountSchema } from '../accounts';

describe('bankAccountSchema', () => {
  it('parses a bank account with all fields', () => {
    const result = bankAccountSchema.parse({
      id: 'BA-1',
      name: 'Business Checking',
      type: 'checking',
      balance: 25000,
      currency: 'USD',
      institution: 'Chase',
      mask: '1234',
    });
    expect(result.id).toBe('BA-1');
    expect(result.name).toBe('Business Checking');
    expect(result.balance).toBe(25000);
    expect(result.currency).toBe('USD');
  });

  it('coerces a numeric id to string', () => {
    const result = bankAccountSchema.parse({ id: 7, name: 'Savings' });
    expect(result.id).toBe('7');
  });

  it('defaults name to empty string when omitted', () => {
    const result = bankAccountSchema.parse({});
    expect(result.name).toBe('');
  });

  it('allows optional fields to be absent', () => {
    const result = bankAccountSchema.parse({ name: 'Checking' });
    expect(result.id).toBeUndefined();
    expect(result.type).toBeUndefined();
    expect(result.balance).toBeUndefined();
  });

  it('allows nullable fields to be null', () => {
    const result = bankAccountSchema.parse({
      name: 'Savings',
      type: null,
      balance: null,
      currency: null,
      institution: null,
      mask: null,
    });
    expect(result.type).toBeNull();
    expect(result.balance).toBeNull();
  });

  it('parses transaction-style fields when present', () => {
    const result = bankAccountSchema.parse({
      name: 'TX',
      matched: true,
      description: 'Coffee',
      date: '2026-05-01',
      category: 'Food',
      amount: -5.5,
    });
    expect(result.matched).toBe(true);
    expect(result.description).toBe('Coffee');
    expect(result.amount).toBe(-5.5);
  });

  it('passes through unknown fields', () => {
    const result = bankAccountSchema.parse({ name: 'Test', extraField: 'extra' });
    expect((result as Record<string, unknown>).extraField).toBe('extra');
  });
});

describe('chartOfAccountSchema', () => {
  it('parses a chart-of-account entry with all fields', () => {
    const result = chartOfAccountSchema.parse({
      id: 'COA-1',
      name: 'Accounts Receivable',
      code: '1200',
      type: 'asset',
      balance: 50000,
    });
    expect(result.id).toBe('COA-1');
    expect(result.name).toBe('Accounts Receivable');
    expect(result.code).toBe('1200');
    expect(result.type).toBe('asset');
    expect(result.balance).toBe(50000);
  });

  it('coerces a numeric id to string', () => {
    const result = chartOfAccountSchema.parse({ id: 5, name: 'Cash' });
    expect(result.id).toBe('5');
  });

  it('defaults name to empty string when omitted', () => {
    const result = chartOfAccountSchema.parse({});
    expect(result.name).toBe('');
  });

  it('allows nullable fields', () => {
    const result = chartOfAccountSchema.parse({
      name: 'Liabilities',
      code: null,
      type: null,
      balance: null,
    });
    expect(result.code).toBeNull();
    expect(result.balance).toBeNull();
  });

  it('passes through unknown fields', () => {
    const result = chartOfAccountSchema.parse({ name: 'Test', description: 'A test account' });
    expect((result as Record<string, unknown>).description).toBe('A test account');
  });
});
