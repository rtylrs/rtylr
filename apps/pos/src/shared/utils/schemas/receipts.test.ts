import { describe, it, expect } from 'vitest';
import { receiptSchema } from './receipts';

describe('receiptSchema', () => {
  const base = {
    id: '1',
    orderId: '100',
    total: 29.99,
    printedAt: '2024-06-01T10:00:00Z',
    items: [{ name: 'Burger', qty: 1, price: 12.5 }],
    payments: [{ id: '1', method: 'cash', amount: 30, createdAt: '2024-06-01T10:00:00Z' }],
  };

  it('parses a fully-populated receipt', () => {
    const result = receiptSchema.parse(base);
    expect(result.id).toBe('1');
    expect(result.orderId).toBe('100');
    expect(result.total).toBe(29.99);
    expect(result.items).toHaveLength(1);
    expect(result.payments).toHaveLength(1);
  });

  it('coerces numeric id to string', () => {
    const result = receiptSchema.parse({ ...base, id: 5 });
    expect(result.id).toBe('5');
  });

  it('coerces numeric orderId to string', () => {
    const result = receiptSchema.parse({ ...base, orderId: 99 });
    expect(result.orderId).toBe('99');
  });

  it('defaults total to 0 when missing', () => {
    const result = receiptSchema.parse({ ...base, total: undefined });
    expect(result.total).toBe(0);
  });

  it('defaults items to an empty array when missing', () => {
    const result = receiptSchema.parse({ ...base, items: undefined });
    expect(result.items).toEqual([]);
  });

  it('defaults payments to an empty array when missing', () => {
    const result = receiptSchema.parse({ ...base, payments: undefined });
    expect(result.payments).toEqual([]);
  });

  it('defaults printedAt to a valid ISO string when missing', () => {
    const result = receiptSchema.parse({ ...base, printedAt: undefined });
    expect(result.printedAt).toMatch(/^\d{4}-\d{2}-\d{2}T/);
  });

  it('falls back to "other" for unknown payment method', () => {
    const result = receiptSchema.parse({
      ...base,
      payments: [{ id: '1', method: 'unknown', amount: 10, createdAt: '2024-06-01T10:00:00Z' }],
    });
    expect(result.payments[0].method).toBe('other');
  });

  it('accepts a customerName field', () => {
    const result = receiptSchema.parse({ ...base, customerName: 'Alice' });
    expect(result.customerName).toBe('Alice');
  });

  it('preserves receipt item name, qty and price', () => {
    const result = receiptSchema.parse(base);
    expect(result.items[0].name).toBe('Burger');
    expect(result.items[0].qty).toBe(1);
    expect(result.items[0].price).toBe(12.5);
  });
});
