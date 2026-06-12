/**
 * End-to-end: load inventory items → perform a stock action →
 * confirm the correct posService endpoint is called with the right payload.
 *
 * Chains buildStockAction (validation + payload construction) →
 * posService.{receiveStock|adjustStock|wasteStock|transferStock}.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    receiveStock: vi.fn(),
    adjustStock: vi.fn(),
    wasteStock: vi.fn(),
    transferStock: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((p) => p),
}));

import { buildStockAction, mapInventoryItems } from './inventoryWorkflow';
import type { StockAction } from './inventoryWorkflow';
import { posService } from '../../shared/services/api/posService';
import { unwrap } from '@voxire/api';

async function applyStockAction(params: {
  itemId: number;
  action: StockAction;
  quantityInput: string;
  branchId: number;
  reason?: string;
}) {
  const { payload, action } = buildStockAction(params);
  switch (action) {
    case 'receive':
      return unwrap(posService.receiveStock(payload));
    case 'adjust':
      return unwrap(posService.adjustStock(payload));
    case 'waste':
      return unwrap(posService.wasteStock(payload));
    case 'transfer':
      return unwrap(posService.transferStock(payload));
  }
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('inventory end-to-end flow', () => {
  it('receive stock calls receiveStock with correct payload', async () => {
    vi.mocked(posService.receiveStock).mockResolvedValue(undefined as never);

    await applyStockAction({ itemId: 10, action: 'receive', quantityInput: '50', branchId: 1, reason: 'delivery' });

    expect(posService.receiveStock).toHaveBeenCalledTimes(1);
    expect(posService.receiveStock).toHaveBeenCalledWith({
      branch_id: 1,
      product_id: 10,
      quantity: 50,
      note: 'delivery',
      employee_id: null,
    });

    expect(posService.adjustStock).not.toHaveBeenCalled();
    expect(posService.wasteStock).not.toHaveBeenCalled();
    expect(posService.transferStock).not.toHaveBeenCalled();
  });

  it('waste stock calls wasteStock with item id and quantity', async () => {
    vi.mocked(posService.wasteStock).mockResolvedValue(undefined as never);

    await applyStockAction({ itemId: 5, action: 'waste', quantityInput: '3', branchId: 1, reason: 'spoiled' });

    expect(posService.wasteStock).toHaveBeenCalledTimes(1);
    expect(posService.wasteStock).toHaveBeenCalledWith({
      branch_id: 1,
      product_id: 5,
      quantity: 3,
      note: 'spoiled',
      employee_id: null,
    });

    expect(posService.receiveStock).not.toHaveBeenCalled();
  });

  it('adjust stock calls adjustStock', async () => {
    vi.mocked(posService.adjustStock).mockResolvedValue(undefined as never);

    await applyStockAction({ itemId: 7, action: 'adjust', quantityInput: '10', branchId: 1, reason: 'count correction' });

    expect(posService.adjustStock).toHaveBeenCalledTimes(1);
    expect(posService.adjustStock).toHaveBeenCalledWith({
      branch_id: 1,
      product_id: 7,
      quantity: 10,
      note: 'count correction',
      employee_id: null,
    });
  });

  it('zero quantity is rejected before any service call', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'receive', quantityInput: '0', branchId: 1 }),
    ).toThrow('Quantity must be greater than zero');

    expect(posService.receiveStock).not.toHaveBeenCalled();
    expect(posService.adjustStock).not.toHaveBeenCalled();
    expect(posService.wasteStock).not.toHaveBeenCalled();
    expect(posService.transferStock).not.toHaveBeenCalled();
  });

  it('mapInventoryItems identifies low-stock items (currentQty < lowStockThreshold)', () => {
    const raw = [
      { item_id: 1, name: 'Milk', current_qty: 2, unit: 'L', low_stock_threshold: 10 },
      { item_id: 2, name: 'Sugar', current_qty: 50, unit: 'kg', low_stock_threshold: 5 },
      { item_id: 3, name: 'Coffee', current_qty: 0, unit: 'kg', low_stock_threshold: 2 },
    ];

    const items = mapInventoryItems(raw);
    expect(items).toHaveLength(3);

    const lowStock = items.filter((i) => i.currentQty < i.lowStockThreshold);
    expect(lowStock).toHaveLength(2);
    expect(lowStock.map((i) => i.name)).toContain('Milk');
    expect(lowStock.map((i) => i.name)).toContain('Coffee');
    expect(lowStock.map((i) => i.name)).not.toContain('Sugar');
  });

  it('low-stock detection uses both snake_case and camelCase fields from server', () => {
    const rawMixed = [
      { item_id: 10, name: 'Oat Milk', current_qty: 1, unit: 'L', low_stock_threshold: 5 },
      { itemId: 11, name: 'Syrup', currentQty: 3, unit: 'bottles', low_stock_threshold: 10 },
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
