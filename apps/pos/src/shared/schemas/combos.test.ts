import { describe, it, expect } from 'vitest';
import { comboSchema } from './combos';

describe('comboSchema', () => {
  const base = { id: '5', name: 'Meal Deal', price: 15.99 };

  it('parses a minimal valid combo', () => {
    const result = comboSchema.parse(base);
    expect(result.id).toBe('5');
    expect(result.name).toBe('Meal Deal');
    expect(result.price).toBe(15.99);
  });

  it('coerces a numeric id to string', () => {
    const result = comboSchema.parse({ ...base, id: 5 });
    expect(result.id).toBe('5');
  });

  it('defaults name to "Combo" when missing', () => {
    const result = comboSchema.parse({ ...base, name: undefined });
    expect(result.name).toBe('Combo');
  });

  it('defaults price to 0 when missing', () => {
    const result = comboSchema.parse({ ...base, price: undefined });
    expect(result.price).toBe(0);
  });

  it('defaults items to an empty array when missing', () => {
    const result = comboSchema.parse(base);
    expect(result.items).toEqual([]);
  });

  it('parses combo items correctly', () => {
    const result = comboSchema.parse({
      ...base,
      items: [{ itemId: '10', qty: 2, name: 'Burger' }],
    });
    expect(result.items).toHaveLength(1);
    expect(result.items[0].itemId).toBe('10');
    expect(result.items[0].qty).toBe(2);
    expect(result.items[0].name).toBe('Burger');
  });

  it('coerces numeric itemId to string in combo items', () => {
    const result = comboSchema.parse({
      ...base,
      items: [{ itemId: 10, qty: 1 }],
    });
    expect(result.items[0].itemId).toBe('10');
  });

  it('defaults item qty to 1 when missing', () => {
    const result = comboSchema.parse({
      ...base,
      items: [{ itemId: '10' }],
    });
    expect(result.items[0].qty).toBe(1);
  });

  it('accepts an optional description', () => {
    const result = comboSchema.parse({ ...base, description: 'Best value' });
    expect(result.description).toBe('Best value');
  });

  it('handles null description gracefully', () => {
    const result = comboSchema.parse({ ...base, description: null });
    expect(result.description).toBeNull();
  });
});
