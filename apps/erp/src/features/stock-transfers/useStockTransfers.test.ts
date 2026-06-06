import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useStockTransfers,
  useCreateStockTransfer,
  useUpdateStockTransferStatus,
} from './useStockTransfers';

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
const sampleTransfers = [
  {
    id: 1,
    ownerId: 1,
    fromWarehouseId: 1,
    toWarehouseId: 2,
    status: 'draft',
    notes: '',
    items: [
      {
        id: 9,
        transferId: 1,
        productId: 5,
        quantity: 10,
        createdAt: '2024-01-01T00:00:00Z',
      },
    ],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('useStockTransfers', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches stock transfers list', async () => {
    mockSafe.get.mockResolvedValue({ stockTransfers: sampleTransfers, hasMore: false });
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.transfers).toHaveLength(1);
    expect(result.current.data?.transfers[0].status).toBe('draft');
  });

  it('uses default pagination params', async () => {
    mockSafe.get.mockResolvedValue({ stockTransfers: [], hasMore: false });
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockSafe.get).toHaveBeenCalledWith('/erp/stock-transfers', {
      params: { page: 1, page_size: 50 },
    });
  });

  it('passes custom pagination params', async () => {
    mockSafe.get.mockResolvedValue({ stockTransfers: [], hasMore: false });
    const { result } = renderHook(() => useStockTransfers({ page: 2, pageSize: 25 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockSafe.get).toHaveBeenCalledWith('/erp/stock-transfers', {
      params: { page: 2, page_size: 25 },
    });
  });

  it('handles API error', async () => {
    mockSafe.get.mockRejectedValue(new Error('fail'));
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('includes hasMore flag in result', async () => {
    mockSafe.get.mockResolvedValue({ stockTransfers: [], hasMore: true });
    const { result } = renderHook(() => useStockTransfers(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.hasMore).toBe(true);
  });
});

describe('useCreateStockTransfer', () => {
  beforeEach(() => vi.clearAllMocks());

  it('posts the backend create payload shape', async () => {
    mockSafe.post.mockResolvedValue({ transferId: 2, message: 'ok' });
    const { result } = renderHook(() => useCreateStockTransfer(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        fromWarehouseId: 1,
        toWarehouseId: 2,
        notes: '',
        items: [{ productId: 5, quantity: 10 }],
      });
    });
    expect(mockSafe.post).toHaveBeenCalledWith('/erp/stock-transfers', {
      fromWarehouseId: 1,
      toWarehouseId: 2,
      notes: '',
      items: [{ productId: 5, quantity: 10 }],
    });
  });
});

describe('useUpdateStockTransferStatus', () => {
  beforeEach(() => vi.clearAllMocks());

  it('puts the status update with lowercase status', async () => {
    mockSafe.put.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useUpdateStockTransferStatus(), {
      wrapper: createWrapper(),
    });
    await act(async () => {
      await result.current.mutateAsync({ id: 1, status: 'in_transit' });
    });
    expect(mockSafe.put).toHaveBeenCalledWith('/erp/stock-transfers/1', {
      id: 1,
      status: 'in_transit',
    });
  });
});
