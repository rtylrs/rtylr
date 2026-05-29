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
});
