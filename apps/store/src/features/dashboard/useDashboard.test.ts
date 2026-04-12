import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useDashboardStats } from './useDashboard';

const mockStoreApi = {
  getDashboardStats: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/storeService', () => ({
  storeApi: {
    getDashboardStats: (...args: any[]) => mockStoreApi.getDashboardStats(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleStats = {
  totalProducts: 120,
  ordersToday: 15,
  revenueToday: 1500,
  activeVisitors: 42,
};

describe('useDashboardStats', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches dashboard stats', async () => {
    mockStoreApi.getDashboardStats.mockResolvedValue(sampleStats);

    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(sampleStats);
  });

  it('calls getDashboardStats', async () => {
    mockStoreApi.getDashboardStats.mockResolvedValue(sampleStats);

    renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockStoreApi.getDashboardStats).toHaveBeenCalledTimes(1));
  });

  it('handles API error', async () => {
    mockStoreApi.getDashboardStats.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('uses store-dashboard query key', async () => {
    mockStoreApi.getDashboardStats.mockResolvedValue(sampleStats);

    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);

    renderHook(() => useDashboardStats(), { wrapper });
    await waitFor(() => expect(client.getQueryData(['store-dashboard'])).toEqual(sampleStats));
  });

  it('returns undefined data while loading', () => {
    mockStoreApi.getDashboardStats.mockReturnValue(new Promise(() => {})); // never resolves

    const { result } = renderHook(() => useDashboardStats(), { wrapper: createWrapper() });
    expect(result.current.data).toBeUndefined();
    expect(result.current.isLoading).toBe(true);
  });
});
