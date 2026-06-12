import { describe, it, expect } from 'vitest';
import { menuItemSchema } from './menuItems';

describe('menuItemSchema', () => {
  const base = { id: '1', name: 'Burger', price: 12.5, category: 'Mains' };

  it('parses a minimal valid menu item', () => {
    const result = menuItemSchema.parse(base);
    expect(result.id).toBe('1');
    expect(result.name).toBe('Burger');
    expect(result.price).toBe(12.5);
    expect(result.category).toBe('Mains');
  });

  it('coerces a numeric id to string', () => {
    const result = menuItemSchema.parse({ ...base, id: 42 });
    expect(result.id).toBe('42');
  });

  it('defaults name to "Item" when missing', () => {
    const result = menuItemSchema.parse({ ...base, name: undefined });
    expect(result.name).toBe('Item');
  });

  it('defaults price to 0 when missing', () => {
    const result = menuItemSchema.parse({ ...base, price: undefined });
    expect(result.price).toBe(0);
  });

  it('defaults category to "Other" when missing', () => {
    const result = menuItemSchema.parse({ ...base, category: undefined });
    expect(result.category).toBe('Other');
  });

  it('defaults available to true when missing', () => {
    const result = menuItemSchema.parse(base);
    expect(result.available).toBe(true);
  });

  it('accepts an explicit available: false', () => {
    const result = menuItemSchema.parse({ ...base, available: false });
    expect(result.available).toBe(false);
  });

  it('maps salePrice alias to price', () => {
    const result = menuItemSchema.parse({ ...base, price: undefined, salePrice: 9.99 });
    expect(result.price).toBe(9.99);
  });

  it('maps category_name alias to category', () => {
    const result = menuItemSchema.parse({ ...base, category: undefined, category_name: 'Drinks' });
    expect(result.category).toBe('Drinks');
  });

  it('maps is_active alias to available', () => {
    const result = menuItemSchema.parse({ ...base, available: undefined, is_active: false });
    expect(result.available).toBe(false);
  });

  it('maps tax_rate alias to taxRate', () => {
    const result = menuItemSchema.parse({ ...base, tax_rate: 0.08 });
    expect(result.taxRate).toBe(0.08);
  });

  it('preserves sku when present', () => {
    const result = menuItemSchema.parse({ ...base, sku: 'BRG-001' });
    expect(result.sku).toBe('BRG-001');
  });

  it('preserves barcode when present', () => {
    const result = menuItemSchema.parse({ ...base, barcode: '123456789' });
    expect(result.barcode).toBe('123456789');
  });

  it('parses a stock value', () => {
    const result = menuItemSchema.parse({ ...base, stock: 42 });
    expect(result.stock).toBe(42);
  });
});
