import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useDashboardStats } from './useDashboardStats';

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

const mockGet = vi.fn();

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleStats = {
  totalProducts: 120,
  lowStockCount: 5,
  pendingOrders: 3,
  totalSuppliers: 12,
  lowStockItems: [{ name: 'Widget', sku: 'WGT-001', stock: 2, minStock: 10, unit: 'pcs' }],
  recentOrders: [],
};

describe('useDashboardStats', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches dashboard stats', async () => {
    mockGet.mockResolvedValue(sampleStats);
    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.totalProducts).toBe(120);
    expect(result.current.data?.lowStockCount).toBe(5);
  });

  it('calls the dashboard endpoint', async () => {
    mockGet.mockResolvedValue(sampleStats);
    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGet).toHaveBeenCalledWith('/erp/dashboard');
  });

  it('handles API error', async () => {
    mockGet.mockRejectedValue(new Error('server error'));
    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('returns low stock items in stats', async () => {
    mockGet.mockResolvedValue(sampleStats);
    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.lowStockItems).toHaveLength(1);
  });

  it('normalizes backend snake_case dashboard payloads', async () => {
    mockGet.mockResolvedValue({
      total_products: 234,
      low_stock_count: 1,
      pending_orders: 2,
      total_suppliers: 3,
      inventory_value: 245561.2,
      low_stock_items: [
        {
          product_id: 9,
          name: 'Creatine',
          sku: 'CRT',
          current_stock: 4,
          reorder_point: 8,
          warehouse_name: 'Main',
        },
      ],
      recent_orders: [
        {
          id: 17,
          po_number: 'PO-17',
          supplier_name: 'North Roast',
          status: 'pending approval',
          total: 120,
          created_at: '2026-06-05T00:00:00Z',
        },
      ],
      top_categories: [{ name: 'Whey', product_count: 35 }],
    });

    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.totalProducts).toBe(234);
    expect(result.current.data?.inventoryValue).toBe(245561.2);
    expect(result.current.data?.lowStockItems[0]).toMatchObject({
      productId: 9,
      stock: 4,
      minStock: 8,
    });
    expect(result.current.data?.recentOrders[0]).toMatchObject({
      poNumber: 'PO-17',
      supplier: 'North Roast',
      status: 'Pending Approval',
      total: 120,
    });
    expect(result.current.data?.topCategories[0]).toEqual({ name: 'Whey', productCount: 35 });
  });
});
