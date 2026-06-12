import { describe, expect, it } from 'vitest';
import { buildStockAction, mapInventoryItems } from './inventoryWorkflow';

describe('buildStockAction', () => {
  it('builds a receive action from valid input', () => {
    const result = buildStockAction({
      itemId: 1,
      action: 'receive',
      quantityInput: '50',
      branchId: 1,
      reason: 'Weekly delivery',
    });

    expect(result).toEqual({
      action: 'receive',
      payload: { branch_id: 1, product_id: 1, quantity: 50, note: 'Weekly delivery', employee_id: null },
    });
  });

  it('builds an adjust action with a fractional quantity', () => {
    const result = buildStockAction({
      itemId: 3,
      action: 'adjust',
      quantityInput: '2.5',
      branchId: 1,
    });

    expect(result.payload.quantity).toBe(2.5);
    expect(result.payload.note).toBeNull();
  });

  it('throws when quantity is zero or negative', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'receive', quantityInput: '0', branchId: 1 })
    ).toThrow('Quantity must be greater than zero.');

    expect(() =>
      buildStockAction({ itemId: 1, action: 'receive', quantityInput: '-5', branchId: 1 })
    ).toThrow('Quantity must be greater than zero.');
  });

  it('throws when quantity is non-numeric', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'receive', quantityInput: 'abc', branchId: 1 })
    ).toThrow('Quantity must be greater than zero.');
  });

  it('throws when item id is invalid', () => {
    expect(() =>
      buildStockAction({ itemId: 0, action: 'receive', quantityInput: '10', branchId: 1 })
    ).toThrow('Select a valid inventory item.');
  });

  it('throws when branch id is missing', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'receive', quantityInput: '10' })
    ).toThrow('Select a terminal location before recording stock movements.');
  });

  it('uses note as reason fallback when reason is absent', () => {
    const result = buildStockAction({
      itemId: 1,
      action: 'waste',
      quantityInput: '3',
      branchId: 1,
      note: 'Spoilage',
    });
    expect(result.payload.note).toBe('Spoilage');
  });
});

describe('mapInventoryItems', () => {
  it('maps snake_case server response', () => {
    const raw = [
      {
        item_id: 1,
        name: 'Tomatoes',
        sku: 'TOM-001',
        current_qty: 15,
        unit: 'kg',
        low_stock_threshold: 5,
      },
    ];
    const result = mapInventoryItems(raw);
    expect(result).toEqual([
      {
        itemId: 1,
        name: 'Tomatoes',
        sku: 'TOM-001',
        currentQty: 15,
        unit: 'kg',
        lowStockThreshold: 5,
      },
    ]);
  });

  it('maps camelCase server response', () => {
    const raw = [{ itemId: 2, name: 'Oil', currentQty: 3, unit: 'L', threshold: 2 }];
    const result = mapInventoryItems(raw);
    expect(result[0].itemId).toBe(2);
    expect(result[0].currentQty).toBe(3);
    expect(result[0].lowStockThreshold).toBe(2);
  });

  it('filters out entries with invalid item ids', () => {
    const raw = [
      { item_id: 0, name: 'Bad', current_qty: 5, unit: 'kg', low_stock_threshold: 1 },
      { item_id: 3, name: 'Good', current_qty: 10, unit: 'kg', low_stock_threshold: 2 },
    ];
    const result = mapInventoryItems(raw);
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe('Good');
  });

  it('returns empty array for non-array input', () => {
    expect(mapInventoryItems(null)).toEqual([]);
    expect(mapInventoryItems('string')).toEqual([]);
    expect(mapInventoryItems({})).toEqual([]);
  });
});
