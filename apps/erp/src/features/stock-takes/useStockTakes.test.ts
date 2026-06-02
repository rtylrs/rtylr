import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useStockTakes, useCreateStockTake, useUpdateStockTake } from './useStockTakes';

const mockSafe = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../../shared/lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockSafe.get(...args),
    post: (...args: unknown[]) => mockSafe.post(...args),
    put: (...args: unknown[]) => mockSafe.put(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

// Response keys are camelCase: @voxire/api converts snake_case responses on the way in.
const sampleStockTakes = [
  {
    id: 1,
    ownerId: 1,
    warehouseId: 1,
    status: 'draft',
    notes: '',
    items: [
      {
        id: 10,
        stockTakeId: 1,
        productId: 5,
        expectedQty: 100,
        countedQty: 0,
        variance: 0,
        createdAt: '2024-01-01T00:00:00Z',
      },
    ],
    startedAt: '',
    completedAt: '',
    createdAt: '2024-01-01T00:00:00Z',
  },
];

describe('useStockTakes', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches stock takes list', async () => {
    mockSafe.get.mockResolvedValue({ stockTakes: sampleStockTakes, hasMore: false });
    const { result } = renderHook(() => useStockTakes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.stockTakes).toHaveLength(1);
    expect(result.current.data?.stockTakes[0].status).toBe('draft');
  });

  it('uses default pagination params', async () => {
    mockSafe.get.mockResolvedValue({ stockTakes: [], hasMore: false });
    const { result } = renderHook(() => useStockTakes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockSafe.get).toHaveBeenCalledWith('/erp/stock-takes', {
      params: { page: 1, page_size: 50 },
    });
  });

  it('passes custom pagination params', async () => {
    mockSafe.get.mockResolvedValue({ stockTakes: [], hasMore: false });
    const { result } = renderHook(() => useStockTakes({ page: 2, pageSize: 10 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockSafe.get).toHaveBeenCalledWith('/erp/stock-takes', {
      params: { page: 2, page_size: 10 },
    });
  });

  it('handles API error', async () => {
    mockSafe.get.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useStockTakes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('includes page and pageSize in result', async () => {
    mockSafe.get.mockResolvedValue({ stockTakes: [], hasMore: true });
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

  it('posts the backend create payload shape', async () => {
    mockSafe.post.mockResolvedValue({ stockTakeId: 2, message: 'ok' });
    const { result } = renderHook(() => useCreateStockTake(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        warehouseId: 1,
        notes: 'Q2 count',
        items: [{ productId: 5, expectedQty: 100, countedQty: 98 }],
      });
    });
    expect(mockSafe.post).toHaveBeenCalledWith('/erp/stock-takes', {
      warehouseId: 1,
      notes: 'Q2 count',
      items: [{ productId: 5, expectedQty: 100, countedQty: 98 }],
    });
  });
});

describe('useUpdateStockTake', () => {
  beforeEach(() => vi.clearAllMocks());

  it('puts the update payload with lowercase status', async () => {
    mockSafe.put.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useUpdateStockTake(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        id: 1,
        status: 'completed',
        items: [{ productId: 5, expectedQty: 100, countedQty: 100 }],
      });
    });
    expect(mockSafe.put).toHaveBeenCalledWith('/erp/stock-takes', {
      id: 1,
      status: 'completed',
      items: [{ productId: 5, expectedQty: 100, countedQty: 100 }],
    });
  });
});
