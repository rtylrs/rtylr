import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useActivity } from './useActivity';

const mockErpApi = {
  listActivity: vi.fn(),
};

vi.mock('@voxire/api', () => ({ unwrap: vi.fn((v) => v) }));

vi.mock('../services/erpApi', () => ({
  erpApi: {
    listActivity: (...args: unknown[]) => mockErpApi.listActivity(...args),
  },
}));

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const sampleActivity = [
  {
    id: 1,
    type: 'product',
    description: 'Product created',
    userId: 1,
    userName: 'Alice',
    createdAt: '2024-01-01T00:00:00Z',
  },
  {
    id: 2,
    type: 'order',
    description: 'Order placed',
    userId: 2,
    userName: 'Bob',
    createdAt: '2024-01-02T00:00:00Z',
  },
];

describe('useActivity', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches activity list with default limit', async () => {
    mockErpApi.listActivity.mockResolvedValue(sampleActivity);
    const { result } = renderHook(() => useActivity(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
    expect(mockErpApi.listActivity).toHaveBeenCalledWith(20);
  });

  it('passes custom limit to API', async () => {
    mockErpApi.listActivity.mockResolvedValue([sampleActivity[0]]);
    const { result } = renderHook(() => useActivity(5), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listActivity).toHaveBeenCalledWith(5);
  });

  it('returns empty array when API returns null', async () => {
    mockErpApi.listActivity.mockResolvedValue(null);
    const { result } = renderHook(() => useActivity(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('handles API error', async () => {
    mockErpApi.listActivity.mockRejectedValue(new Error('network error'));
    const { result } = renderHook(() => useActivity(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('includes limit in query key', async () => {
    mockErpApi.listActivity.mockResolvedValue([]);
    const { result } = renderHook(() => useActivity(10), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockErpApi.listActivity).toHaveBeenCalledWith(10);
  });
});
