/**
 * End-to-end: load inventory items → perform a stock action →
 * confirm the correct posService endpoint is called with the right payload.
 *
 * Chains buildStockAction (validation + payload construction) →
 * posService.{receiveStock|adjustStock|wasteStock|transferStock}.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../shared/services/posService', () => ({
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
import { posService } from '../../shared/services/posService';
import { unwrap } from '@voxire/api';

// Helper that mirrors what the inventory UI does: build → dispatch.
async function applyStockAction(params: {
  itemId: number;
  action: StockAction;
  quantityInput: string;
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

    await applyStockAction({ itemId: 10, action: 'receive', quantityInput: '50', reason: 'delivery' });

    expect(posService.receiveStock).toHaveBeenCalledTimes(1);
    expect(posService.receiveStock).toHaveBeenCalledWith({
      itemId: 10,
      quantity: 50,
      reason: 'delivery',
    });

    expect(posService.adjustStock).not.toHaveBeenCalled();
    expect(posService.wasteStock).not.toHaveBeenCalled();
    expect(posService.transferStock).not.toHaveBeenCalled();
  });

  it('waste stock calls wasteStock with item id and quantity', async () => {
    vi.mocked(posService.wasteStock).mockResolvedValue(undefined as never);

    await applyStockAction({ itemId: 5, action: 'waste', quantityInput: '3', reason: 'spoiled' });

    expect(posService.wasteStock).toHaveBeenCalledTimes(1);
    expect(posService.wasteStock).toHaveBeenCalledWith({
      itemId: 5,
      quantity: 3,
      reason: 'spoiled',
    });

    expect(posService.receiveStock).not.toHaveBeenCalled();
  });

  it('adjust stock calls adjustStock', async () => {
    vi.mocked(posService.adjustStock).mockResolvedValue(undefined as never);

    await applyStockAction({ itemId: 7, action: 'adjust', quantityInput: '10', reason: 'count correction' });

    expect(posService.adjustStock).toHaveBeenCalledTimes(1);
    expect(posService.adjustStock).toHaveBeenCalledWith({
      itemId: 7,
      quantity: 10,
      reason: 'count correction',
    });
  });

  it('zero quantity is rejected before any service call', () => {
    expect(() =>
      buildStockAction({ itemId: 1, action: 'receive', quantityInput: '0' }),
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
    // Milk (2 < 10) and Coffee (0 < 2) are low; Sugar (50 >= 5) is not.
    expect(lowStock).toHaveLength(2);
    expect(lowStock.map((i) => i.name)).toContain('Milk');
    expect(lowStock.map((i) => i.name)).toContain('Coffee');
    expect(lowStock.map((i) => i.name)).not.toContain('Sugar');
  });

  it('low-stock detection uses both snake_case id/qty fields and camelCase id/qty fields from server', () => {
    const rawMixed = [
      // snake_case id + qty + threshold (typical server response)
      { item_id: 10, name: 'Oat Milk', current_qty: 1, unit: 'L', low_stock_threshold: 5 },
      // camelCase id + qty with snake_case threshold (hybrid — clients that normalise some fields)
      { itemId: 11, name: 'Syrup', currentQty: 3, unit: 'bottles', low_stock_threshold: 10 },
    ];

    const items = mapInventoryItems(rawMixed);
    expect(items).toHaveLength(2);

    const [oatMilk, syrup] = items;

    // snake_case fields resolved correctly
    expect(oatMilk.itemId).toBe(10);
    expect(oatMilk.currentQty).toBe(1);
    expect(oatMilk.lowStockThreshold).toBe(5);
    expect(oatMilk.currentQty).toBeLessThan(oatMilk.lowStockThreshold);

    // camelCase id + qty resolved correctly, threshold via snake_case key
    expect(syrup.itemId).toBe(11);
    expect(syrup.currentQty).toBe(3);
    expect(syrup.lowStockThreshold).toBe(10);
    expect(syrup.currentQty).toBeLessThan(syrup.lowStockThreshold);
  });
});
