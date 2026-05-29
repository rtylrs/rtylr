import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { usePriceLists, useCreatePriceList } from './usePriceLists';

const mockErpApi = {
  listPriceLists: vi.fn(),
  createPriceList: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listPriceLists: (...args: unknown[]) => mockErpApi.listPriceLists(...args),
    createPriceList: (...args: unknown[]) => mockErpApi.createPriceList(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const samplePriceLists = [
  {
    id: 1,
    name: 'Retail',
    description: 'Standard retail',
    tier: 'Retail' as const,
    multiplier: 1.0,
    activeCount: 100,
  },
  {
    id: 2,
    name: 'Wholesale',
    description: 'Wholesale pricing',
    tier: 'Wholesale' as const,
    multiplier: 0.8,
    activeCount: 50,
  },
];

describe('usePriceLists', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches price lists', async () => {
    mockErpApi.listPriceLists.mockResolvedValue(samplePriceLists);
    const { result } = renderHook(() => usePriceLists(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when API returns null', async () => {
    mockErpApi.listPriceLists.mockResolvedValue(null);
    const { result } = renderHook(() => usePriceLists(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockErpApi.listPriceLists.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => usePriceLists(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('uses "price-lists" query key', async () => {
    mockErpApi.listPriceLists.mockResolvedValue([]);
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);
    renderHook(() => usePriceLists(), { wrapper });
    await waitFor(() => expect(client.getQueryData(['price-lists'])).toEqual([]));
  });
});

describe('useCreatePriceList', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createPriceList with data', async () => {
    const newList = {
      id: 3,
      name: 'VIP',
      description: '',
      tier: 'VIP' as const,
      multiplier: 0.7,
      activeCount: 0,
    };
    mockErpApi.createPriceList.mockResolvedValue(newList);
    const { result } = renderHook(() => useCreatePriceList(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ name: 'VIP', tier: 'VIP', multiplier: 0.7 });
    });
    expect(mockErpApi.createPriceList).toHaveBeenCalledWith({
      name: 'VIP',
      tier: 'VIP',
      multiplier: 0.7,
    });
  });
});
