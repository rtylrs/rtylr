import { describe, it, expect } from 'vitest';
import { kdsOrderSchema } from './kdsOrders';

describe('kdsOrderSchema', () => {
  const base = {
    id: '1',
    station: 'Kitchen',
    status: 'new',
    createdAt: '2024-06-01T10:00:00Z',
    items: [{ name: 'Burger', qty: 2 }],
  };

  it('parses a minimal valid KDS order', () => {
    const result = kdsOrderSchema.parse(base);
    expect(result.id).toBe('1');
    expect(result.station).toBe('Kitchen');
    expect(result.status).toBe('new');
  });

  it('coerces a numeric id to string', () => {
    const result = kdsOrderSchema.parse({ ...base, id: 42 });
    expect(result.id).toBe('42');
  });

  it('defaults station to "Kitchen" when missing', () => {
    const result = kdsOrderSchema.parse({ ...base, station: undefined });
    expect(result.station).toBe('Kitchen');
  });

  it('defaults status to "new" for unknown status values', () => {
    const result = kdsOrderSchema.parse({ ...base, status: 'unknown' });
    expect(result.status).toBe('new');
  });

  it('accepts all valid status values', () => {
    for (const status of ['new', 'cooking', 'ready', 'served'] as const) {
      const result = kdsOrderSchema.parse({ ...base, status });
      expect(result.status).toBe(status);
    }
  });

  it('defaults items to empty array when missing', () => {
    const result = kdsOrderSchema.parse({ ...base, items: undefined });
    expect(result.items).toEqual([]);
  });

  it('parses items with name and qty', () => {
    const result = kdsOrderSchema.parse(base);
    expect(result.items[0].name).toBe('Burger');
    expect(result.items[0].qty).toBe(2);
  });

  it('defaults item name to "Item" when missing', () => {
    const result = kdsOrderSchema.parse({ ...base, items: [{ qty: 1 }] });
    expect(result.items[0].name).toBe('Item');
  });

  it('defaults item qty to 1 when missing', () => {
    const result = kdsOrderSchema.parse({ ...base, items: [{ name: 'Fries' }] });
    expect(result.items[0].qty).toBe(1);
  });

  it('preserves optional course and notes fields', () => {
    const result = kdsOrderSchema.parse({
      ...base,
      items: [{ name: 'Soup', qty: 1, course: 'Starter', notes: 'No salt' }],
    });
    expect(result.items[0].course).toBe('Starter');
    expect(result.items[0].notes).toBe('No salt');
  });

  it('accepts an optional order number', () => {
    const result = kdsOrderSchema.parse({ ...base, number: '42' });
    expect(result.number).toBe('42');
  });

  it('accepts an optional elapsedSeconds', () => {
    const result = kdsOrderSchema.parse({ ...base, elapsedSeconds: 120 });
    expect(result.elapsedSeconds).toBe(120);
  });
});
