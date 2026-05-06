import { describe, expect, it } from 'vitest';
import { buildMenuItemUpdate } from './productQuickEditWorkflow';

describe('buildMenuItemUpdate', () => {
  it('builds a normalized menu item update', () => {
    expect(
      buildMenuItemUpdate('42.9', {
        price: 12.5,
        stock: 8.9,
        available: true,
      })
    ).toEqual({
      itemId: 42,
      payload: { price: 12.5, stock: 8, available: true },
    });
  });

  it('rejects invalid menu item ids', () => {
    expect(() =>
      buildMenuItemUpdate('local-temp', {
        price: 12.5,
        available: true,
      })
    ).toThrow('Select a valid product before saving.');
  });

  it('normalizes invalid price and stock values', () => {
    expect(
      buildMenuItemUpdate('42', {
        price: Number.NaN,
        stock: -4,
        available: false,
      })
    ).toEqual({
      itemId: 42,
      payload: { price: 0, stock: 0, available: false },
    });
  });

  it('preserves omitted stock', () => {
    expect(
      buildMenuItemUpdate('42', {
        price: 9,
        stock: undefined,
        available: true,
      })
    ).toEqual({
      itemId: 42,
      payload: { price: 9, available: true },
    });
  });
});
