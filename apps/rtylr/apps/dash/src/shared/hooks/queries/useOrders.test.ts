import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  put: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => value,
}));

vi.mock('../../lib/clients', () => ({
  rest: { get: mocks.get, put: mocks.put },
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

import { useOrders } from './useOrders';

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useOrders', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.get.mockResolvedValue({
      orders: [
        {
          id: '1',
          status: 'pending',
          subtotal: 10,
          total: 10,
          items: [{ id: 'i1', name: 'Coffee', quantity: 1, price: 10, total: 9 }],
        },
      ],
      total: 51,
      has_more: true,
      page: 1,
      page_size: 50,
    });
  });

  it('uses the registered dashboard list route and preserves its pagination metadata', async () => {
    mocks.get.mockResolvedValueOnce({
      orders: [
        {
          id: '1',
          status: 'pending',
          subtotal: 10,
          total: 10,
          items: [{ id: 'i1', name: 'Coffee', quantity: 1, price: 10, total: 9 }],
        },
      ],
      total: 51,
      has_more: true,
      page: 3,
      page_size: 25,
    });

    const { result } = renderHook(() => useOrders({ limit: 25, offset: 50, status: 'pending' }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.get).toHaveBeenCalledWith('/pos/orders?page=3&page_size=25&status=pending');
    expect(result.current.data).toMatchObject({
      total: 51,
      hasMore: true,
      page: 3,
      pageSize: 25,
    });
    expect(result.current.data?.orders[0].items?.[0].total).toBe(9);
  });
});
