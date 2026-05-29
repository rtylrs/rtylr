import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useStockTakes, useCreateStockTake } from './useStockTakes';

const mockErpApi = {
  listStockTakes: vi.fn(),
  createStockTake: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listStockTakes: (...args: unknown[]) => mockErpApi.listStockTakes(...args),
    createStockTake: (...args: unknown[]) => mockErpApi.createStockTake(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleStockTakes = [
  {
    id: 1,
    name: 'Q1 Count',
    warehouseId: 1,
    warehouseName: 'Main',
    status: 'Draft',
    itemsCounted: 0,
    totalItems: 100,
    variance: 0,
    createdAt: '2024-01-01T00:00:00Z',
  },
];

describe('useStockTakes', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches stock takes list', async () => {
    mockErpApi.listStockTakes.mockResolvedValue({ stockTakes: sampleStockTakes, has_more: false });
    const { result } = renderHook(() => useStockTakes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.stockTakes).toHaveLength(1);
  });

  it('uses default pagination params', async () => {
    mockErpApi.listStockTakes.mockResolvedValue({ stockTakes: [], has_more: false });
    const { result } = renderHook(() => useStockTakes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listStockTakes).toHaveBeenCalledWith({ page: 1, page_size: 50 });
  });

  it('passes custom pagination params', async () => {
    mockErpApi.listStockTakes.mockResolvedValue({ stockTakes: [], has_more: false });
    const { result } = renderHook(() => useStockTakes({ page: 2, pageSize: 10 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listStockTakes).toHaveBeenCalledWith({ page: 2, page_size: 10 });
  });

  it('handles API error', async () => {
    mockErpApi.listStockTakes.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useStockTakes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('includes page and pageSize in result', async () => {
    mockErpApi.listStockTakes.mockResolvedValue({ stockTakes: [], has_more: true });
    const { result } = renderHook(() => useStockTakes({ page: 3, pageSize: 20 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.page).toBe(3);
    expect(result.current.data?.pageSize).toBe(20);
    expect(result.current.data?.hasMore).toBe(true);
  });
});

describe('useCreateStockTake', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createStockTake with data', async () => {
    mockErpApi.createStockTake.mockResolvedValue({ id: 2, name: 'New Take' });
    const { result } = renderHook(() => useCreateStockTake(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ name: 'New Take', warehouseId: 1 });
    });
    expect(mockErpApi.createStockTake).toHaveBeenCalledWith({ name: 'New Take', warehouseId: 1 });
  });
});
