/**
 * End-to-end: load inventory items → build generated stock action payloads.
 */
import { describe, it, expect } from 'vitest';

import { buildStockAction, mapInventoryItems } from './inventoryWorkflow';
import type { StockAction } from './inventoryWorkflow';

function applyStockAction(params: {
  itemId: number;
  action: StockAction;
  quantityInput: string;
  branchId: number;
  toBranchId?: number;
  reason?: string;
}) {
  return buildStockAction(params);
}

describe('inventory end-to-end flow', () => {
  it('receive stock builds the generated payload', () => {
    expect(
      applyStockAction({ itemId: 10, action: 'receive', quantityInput: '50', branchId: 1, reason: 'delivery' }),
    ).toEqual({
      action: 'receive',
      payload: {
      branchId: '1',
      productId: '10',
      quantity: 50,
      note: 'delivery',
      employeeId: '0',
      },
    });
  });

  it('waste stock builds the generated payload', () => {
    expect(
      applyStockAction({ itemId: 5, action: 'waste', quantityInput: '3', branchId: 1, reason: 'spoiled' }),
    ).toEqual({
      action: 'waste',
      payload: {
      branchId: '1',
      productId: '5',
      quantity: 3,
      note: 'spoiled',
      employeeId: '0',
      },
    });
  });

  it('adjust stock builds the generated payload', () => {
    expect(
      applyStockAction({ itemId: 7, action: 'adjust', quantityInput: '10', branchId: 1, reason: 'count correction' }),
    ).toEqual({
      action: 'adjust',
      payload: {
      branchId: '1',
      productId: '7',
      quantity: 10,
      note: 'count correction',
      employeeId: '0',
      },
    });
  });

  it('transfer stock builds the generated payload with source and destination branches', () => {
    expect(applyStockAction({
      itemId: 7,
      action: 'transfer',
      quantityInput: '5',
      branchId: 1,
      toBranchId: 2,
      reason: 'restock front counter',
    })).toEqual({
      action: 'transfer',
      payload: {
      fromBranchId: '1',
      toBranchId: '2',
      productId: '7',
      quantity: 5,
      note: 'restock front counter',
      employeeId: '0',
      },
    });
  });

  it('transfer stock requires a destination branch', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'transfer', quantityInput: '1', branchId: 1 }),
    ).toThrow('Select a destination location before transferring stock');
  });

  it('zero quantity is rejected before any service call', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'receive', quantityInput: '0', branchId: 1 }),
    ).toThrow('Quantity must be greater than zero');
  });

  it('mapInventoryItems identifies low-stock items (currentQty < lowStockThreshold)', () => {
    const raw = [
      { id: '1', businessId: '1', branchId: '1', productId: '1', itemId: '1', name: 'Milk', sku: '', quantity: 2, currentQty: 2, unit: 'L', reorderPoint: 10, reorderQuantity: 0, lowStockThreshold: 10, updatedAt: '' },
      { id: '2', businessId: '1', branchId: '1', productId: '2', itemId: '2', name: 'Sugar', sku: '', quantity: 50, currentQty: 50, unit: 'kg', reorderPoint: 5, reorderQuantity: 0, lowStockThreshold: 5, updatedAt: '' },
      { id: '3', businessId: '1', branchId: '1', productId: '3', itemId: '3', name: 'Coffee', sku: '', quantity: 0, currentQty: 0, unit: 'kg', reorderPoint: 2, reorderQuantity: 0, lowStockThreshold: 2, updatedAt: '' },
    ];

    const items = mapInventoryItems(raw);
    expect(items).toHaveLength(3);

    const lowStock = items.filter((i) => i.currentQty < i.lowStockThreshold);
    expect(lowStock).toHaveLength(2);
    expect(lowStock.map((i) => i.name)).toContain('Milk');
    expect(lowStock.map((i) => i.name)).toContain('Coffee');
    expect(lowStock.map((i) => i.name)).not.toContain('Sugar');
  });

  it('low-stock detection uses generated inventory fields', () => {
    const rawMixed = [
      { id: '10', businessId: '1', branchId: '1', productId: '10', itemId: '10', name: 'Oat Milk', sku: '', quantity: 1, currentQty: 1, unit: 'L', reorderPoint: 5, reorderQuantity: 0, lowStockThreshold: 5, updatedAt: '' },
      { id: '11', businessId: '1', branchId: '1', productId: '11', itemId: '11', name: 'Syrup', sku: '', quantity: 3, currentQty: 3, unit: 'bottles', reorderPoint: 10, reorderQuantity: 0, lowStockThreshold: 10, updatedAt: '' },
    ];

    const items = mapInventoryItems(rawMixed);
    expect(items).toHaveLength(2);

    const [oatMilk, syrup] = items;

    expect(oatMilk.itemId).toBe(10);
    expect(oatMilk.currentQty).toBe(1);
    expect(oatMilk.lowStockThreshold).toBe(5);
    expect(oatMilk.currentQty).toBeLessThan(oatMilk.lowStockThreshold);

    expect(syrup.itemId).toBe(11);
    expect(syrup.currentQty).toBe(3);
    expect(syrup.lowStockThreshold).toBe(10);
    expect(syrup.currentQty).toBeLessThan(syrup.lowStockThreshold);
  });
});
