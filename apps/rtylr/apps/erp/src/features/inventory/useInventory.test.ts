import { beforeEach, describe, expect, it, vi } from 'vitest';

const inventoryClient = vi.hoisted(() => ({
  getInventoryReport: vi.fn(),
  getLowStock: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => inventoryClient,
}));

import { fetchAllInventoryReport, fetchAllLowStock } from './useInventory';

describe('inventory pagination', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uses report metadata to fetch inventory beyond the first page', async () => {
    inventoryClient.getInventoryReport
      .mockResolvedValueOnce({
        items: Array.from({ length: 100 }, (_, index) => ({ productId: `p-${index}` })),
        total: 150,
      })
      .mockResolvedValueOnce({
        items: Array.from({ length: 50 }, (_, index) => ({ productId: `p-${100 + index}` })),
        total: 150,
        hasMore: false,
      });

    const items = await fetchAllInventoryReport();

    expect(items).toHaveLength(150);
    expect(inventoryClient.getInventoryReport).toHaveBeenNthCalledWith(1, {
      limit: 100,
      offset: 0,
    });
    expect(inventoryClient.getInventoryReport).toHaveBeenNthCalledWith(2, {
      limit: 100,
      offset: 100,
    });
  });

  it('fetches every low-stock page', async () => {
    inventoryClient.getLowStock
      .mockResolvedValueOnce({
        items: Array.from({ length: 100 }, (_, index) => ({ productId: `low-${index}` })),
        total: 150,
        hasMore: true,
      })
      .mockResolvedValueOnce({
        items: Array.from({ length: 50 }, (_, index) => ({ productId: `low-${100 + index}` })),
        total: 150,
        hasMore: false,
      });

    const items = await fetchAllLowStock();

    expect(items).toHaveLength(150);
    expect(inventoryClient.getLowStock).toHaveBeenNthCalledWith(2, {
      limit: 100,
      offset: 100,
    });
  });
});
