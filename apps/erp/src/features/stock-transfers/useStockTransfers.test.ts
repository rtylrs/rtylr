import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useStockTransfers, useCreateStockTransfer } from './useStockTransfers';

const mockErpApi = {
  listStockTransfers: vi.fn(),
  createStockTransfer: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/services/erpApi', () => ({
  erpApi: {
    listStockTransfers: (...args: unknown[]) => mockErpApi.listStockTransfers(...args),
    createStockTransfer: (...args: unknown[]) => mockErpApi.createStockTransfer(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleTransfers = [
  {
    id: 1,
    fromWarehouseId: 1,
    toWarehouseId: 2,
    fromName: 'Main',
    toName: 'Secondary',
    items: 5,
    status: 'Draft',
    createdAt: '2024-01-01T00:00:00Z',
  },
];

describe('useStockTransfers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches stock transfers list', async () => {
    mockErpApi.listStockTransfers.mockResolvedValue({
      stockTransfers: sampleTransfers,
      has_more: false,
    });
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.transfers).toHaveLength(1);
  });

  it('uses default pagination params', async () => {
    mockErpApi.listStockTransfers.mockResolvedValue({ stockTransfers: [], has_more: false });
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listStockTransfers).toHaveBeenCalledWith({ page: 1, page_size: 50 });
  });

  it('passes custom pagination params', async () => {
    mockErpApi.listStockTransfers.mockResolvedValue({ stockTransfers: [], has_more: false });
    const { result } = renderHook(() => useStockTransfers({ page: 2, pageSize: 25 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listStockTransfers).toHaveBeenCalledWith({ page: 2, page_size: 25 });
  });

  it('handles API error', async () => {
    mockErpApi.listStockTransfers.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('includes hasMore flag in result', async () => {
    mockErpApi.listStockTransfers.mockResolvedValue({ stockTransfers: [], has_more: true });
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.hasMore).toBe(true);
  });
});

describe('useCreateStockTransfer', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls createStockTransfer with data', async () => {
    mockErpApi.createStockTransfer.mockResolvedValue({
      id: 2,
      fromWarehouseId: 1,
      toWarehouseId: 2,
    });
    const { result } = renderHook(() => useCreateStockTransfer(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({ fromWarehouseId: 1, toWarehouseId: 2 });
    });
    expect(mockErpApi.createStockTransfer).toHaveBeenCalledWith({
      fromWarehouseId: 1,
      toWarehouseId: 2,
    });
  });
});
