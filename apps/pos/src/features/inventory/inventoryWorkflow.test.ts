import { describe, expect, it } from 'vitest';
import { buildStockAction, mapInventoryItems, mapInventoryLocations } from './inventoryWorkflow';

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

  it('builds a transfer action with source and destination branches', () => {
    const result = buildStockAction({
      itemId: 12,
      action: 'transfer',
      quantityInput: '4',
      branchId: 1,
      toBranchId: 2,
      reason: 'Move to warehouse',
      employeeId: 9,
    });

    expect(result).toEqual({
      action: 'transfer',
      payload: {
        from_branch_id: 1,
        to_branch_id: 2,
        product_id: 12,
        quantity: 4,
        note: 'Move to warehouse',
        employee_id: 9,
      },
    });
  });

  it('rejects transfer without a distinct destination branch', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'transfer', quantityInput: '2', branchId: 1 }),
    ).toThrow('Select a destination location before transferring stock.');

    expect(() =>
      buildStockAction({
        itemId: 1,
        action: 'transfer',
        quantityInput: '2',
        branchId: 1,
        toBranchId: 1,
      }),
    ).toThrow('Destination location must be different from the current location.');
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

  it('maps shared stock inventory response', () => {
    const raw = [
      {
        id: 99,
        product_id: 12,
        name: 'Platinum MultiVitamins Muscletech',
        sku: 'SUPLAB-PLATINUM-MULTIVITAMINS',
        quantity: 28,
        unit: 'units',
        reorder_point: 5,
      },
    ];
    const result = mapInventoryItems(raw);
    expect(result[0]).toEqual({
      itemId: 12,
      name: 'Platinum MultiVitamins Muscletech',
      sku: 'SUPLAB-PLATINUM-MULTIVITAMINS',
      currentQty: 28,
      unit: 'units',
      lowStockThreshold: 5,
    });
  });

  it('maps inventory envelope response', () => {
    const result = mapInventoryItems({
      inventory: [{ product_id: 7, name: 'Iso 100 Dymatize', quantity: 8 }],
    });
    expect(result[0].itemId).toBe(7);
    expect(result[0].currentQty).toBe(8);
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

describe('mapInventoryLocations', () => {
  it('maps location envelope responses', () => {
    expect(
      mapInventoryLocations({
        locations: [
          { id: 1, name: 'Main' },
          { location_id: 2, label: 'Warehouse' },
        ],
      }),
    ).toEqual([
      { id: 1, name: 'Main' },
      { id: 2, name: 'Warehouse' },
    ]);
  });

  it('filters invalid locations and falls back to generated names', () => {
    expect(mapInventoryLocations([{ id: 0, name: 'Bad' }, { locationId: 3 }])).toEqual([
      { id: 3, name: 'Location 3' },
    ]);
  });
});
