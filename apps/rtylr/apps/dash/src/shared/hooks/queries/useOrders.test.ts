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
      orders: [{ id: '1', status: 'pending', subtotal: 10, total: 10 }],
    });
  });

  it('uses the registered dashboard list route with page pagination', async () => {
    const { result } = renderHook(() => useOrders({ limit: 25, offset: 50, status: 'pending' }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.get).toHaveBeenCalledWith('/pos/orders?page=3&page_size=25&status=pending');
  });
});
