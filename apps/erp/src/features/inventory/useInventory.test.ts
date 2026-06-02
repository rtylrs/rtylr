import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useInventoryReport, useLowStock, useAdjustInventory } from './useInventory';

const mockErpApi = {
  getInventoryReport: vi.fn(),
  getLowStock: vi.fn(),
};

const mockSafe = {
  post: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    getInventoryReport: (...args: unknown[]) => mockErpApi.getInventoryReport(...args),
    getLowStock: (...args: unknown[]) => mockErpApi.getLowStock(...args),
  },
}));

vi.mock('../../shared/lib/api', () => ({
  safe: {
    post: (...args: unknown[]) => mockSafe.post(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useInventoryReport', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches inventory report', async () => {
    const report = {
      items: [
        {
          productId: 1,
          productName: 'Widget A',
          sku: 'WA-001',
          unit: 'pcs',
          stockQuantity: 100,
          lowStockThreshold: 10,
          reorderQuantity: 50,
          costPrice: 5,
          salePrice: 8,
          stockValue: 500,
          isLowStock: false,
        },
        {
          productId: 2,
          productName: 'Widget B',
          sku: 'WB-001',
          unit: 'pcs',
          stockQuantity: 3,
          lowStockThreshold: 10,
          reorderQuantity: 50,
          costPrice: 10,
          salePrice: 15,
          stockValue: 30,
          isLowStock: true,
        },
      ],
      totalValue: 530,
    };
    mockErpApi.getInventoryReport.mockResolvedValue(report);

    const { result } = renderHook(() => useInventoryReport(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
  });

  it('handles API error', async () => {
    mockErpApi.getInventoryReport.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useInventoryReport(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('uses inventory-report query key', async () => {
    mockErpApi.getInventoryReport.mockResolvedValue([]);

    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);

    renderHook(() => useInventoryReport(), { wrapper });
    await waitFor(() => expect(client.getQueryData(['inventory-report'])).toEqual([]));
  });
});

describe('useLowStock', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches low stock items', async () => {
    const items = {
      items: [
        {
          productId: 2,
          productName: 'Widget B',
          sku: 'WB-001',
          unit: 'pcs',
          stockQuantity: 3,
          lowStockThreshold: 10,
          reorderQuantity: 50,
          costPrice: 10,
          salePrice: 15,
          stockValue: 30,
          isLowStock: true,
        },
      ],
      count: 1,
    };
    mockErpApi.getLowStock.mockResolvedValue(items);

    const { result } = renderHook(() => useLowStock(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
  });

  it('handles API error', async () => {
    mockErpApi.getLowStock.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useLowStock(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useAdjustInventory', () => {
  beforeEach(() => vi.clearAllMocks());

  it('posts the backend adjust body to /erp/inventory/adjust', async () => {
    mockSafe.post.mockResolvedValue({ message: 'Inventory adjusted successfully' });

    const { result } = renderHook(() => useAdjustInventory(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        productId: 1,
        change: 10,
        type: 'purchase',
        notes: 'Restock',
      });
    });

    expect(mockSafe.post).toHaveBeenCalledWith('/erp/inventory/adjust', {
      productId: 1,
      change: 10,
      type: 'purchase',
      notes: 'Restock',
    });
  });

  it('handles mutation error', async () => {
    mockSafe.post.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useAdjustInventory(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({ productId: 1, change: -5, type: 'waste' });
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});
