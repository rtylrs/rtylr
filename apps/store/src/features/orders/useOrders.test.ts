import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useOrders, useOrder, useUpdateOrderStatus } from './useOrders';

const mockStoreApi = {
  listOrders: vi.fn(),
  getOrder: vi.fn(),
  updateOrderStatus: vi.fn(),
};

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../shared/services/storeService', () => ({
  storeApi: {
    listOrders: (...args: unknown[]) => mockStoreApi.listOrders(...args),
    getOrder: (...args: unknown[]) => mockStoreApi.getOrder(...args),
    updateOrderStatus: (...args: unknown[]) => mockStoreApi.updateOrderStatus(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleOrders = [
  {
    id: 1,
    orderNumber: 'ORD-001',
    customerName: 'Alice',
    customerEmail: 'alice@test.com',
    status: 'pending',
    total: 99.99,
    subtotal: 89.99,
    shippingCost: 5.0,
    tax: 5.0,
    itemCount: 2,
    paymentMethod: 'card',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
  {
    id: 2,
    orderNumber: 'ORD-002',
    customerName: 'Bob',
    customerEmail: 'bob@test.com',
    status: 'shipped',
    total: 49.99,
    subtotal: 44.99,
    shippingCost: 3.0,
    tax: 2.0,
    itemCount: 1,
    paymentMethod: 'card',
    createdAt: '2026-01-02T00:00:00Z',
    updatedAt: '2026-01-02T00:00:00Z',
  },
];

describe('useOrders', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches orders without filters', async () => {
    mockStoreApi.listOrders.mockResolvedValue(sampleOrders);

    const { result } = renderHook(() => useOrders(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
  });

  it('passes filters to listOrders', async () => {
    mockStoreApi.listOrders.mockResolvedValue([]);

    const filters = { search: 'alice', status: 'pending' as const, limit: 5 };
    renderHook(() => useOrders(filters), { wrapper: createWrapper() });

    await waitFor(() =>
      expect(mockStoreApi.listOrders).toHaveBeenCalledWith({
        search: 'alice',
        status: 'pending',
        limit: 5,
        offset: undefined,
      })
    );
  });

  it('handles API error', async () => {
    mockStoreApi.listOrders.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useOrders(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('separates cache by filters', async () => {
    mockStoreApi.listOrders.mockResolvedValue([]);

    const w = createWrapper();
    renderHook(() => useOrders({ status: 'pending' as const }), { wrapper: w });
    renderHook(() => useOrders({ status: 'shipped' as const }), { wrapper: w });

    await waitFor(() => expect(mockStoreApi.listOrders).toHaveBeenCalledTimes(2));
  });
});

describe('useOrder', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single order by id', async () => {
    mockStoreApi.getOrder.mockResolvedValue(sampleOrders[0]);

    const { result } = renderHook(() => useOrder(1), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.orderNumber).toBe('ORD-001');
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHook(() => useOrder(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useUpdateOrderStatus', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls updateOrderStatus with correct args', async () => {
    mockStoreApi.updateOrderStatus.mockResolvedValue({ success: true });

    const { result } = renderHook(() => useUpdateOrderStatus(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({ id: 1, status: 'shipped' as const });
    });

    expect(mockStoreApi.updateOrderStatus).toHaveBeenCalledWith(1, 'shipped');
  });

  it('handles mutation error', async () => {
    mockStoreApi.updateOrderStatus.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useUpdateOrderStatus(), { wrapper: createWrapper() });

    await act(async () => {
      result.current.mutate({ id: 1, status: 'shipped' as const });
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
  });
});
