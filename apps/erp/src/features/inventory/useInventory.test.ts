import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useInventoryReport, useLowStock, useAdjustInventory } from './useInventory';

const mockErpApi = {
  getInventoryReport: vi.fn(),
  getLowStock: vi.fn(),
  adjustInventory: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    getInventoryReport: (...args: any[]) => mockErpApi.getInventoryReport(...args),
    getLowStock: (...args: any[]) => mockErpApi.getLowStock(...args),
    adjustInventory: (...args: any[]) => mockErpApi.adjustInventory(...args),
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
    const report = [
      { productId: 1, name: 'Widget A', sku: 'WA-001', currentStock: 100, value: 500 },
      { productId: 2, name: 'Widget B', sku: 'WB-001', currentStock: 3, value: 30 },
    ];
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
    const items = [{ name: 'Widget B', sku: 'WB-001', stock: 3, minStock: 10 }];
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

  it('calls adjustInventory with data', async () => {
    mockErpApi.adjustInventory.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useAdjustInventory(), { wrapper: createWrapper() });

    const adjustData = { productId: 1, adjustment: 10, reason: 'Restock' };
    await act(async () => {
      await result.current.mutateAsync(adjustData);
    });

    expect(mockErpApi.adjustInventory).toHaveBeenCalledWith(adjustData);
  });

  it('handles mutation error', async () => {
    mockErpApi.adjustInventory.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useAdjustInventory(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({ productId: 1, adjustment: -5 });
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});
