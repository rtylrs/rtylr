// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/lib/api', () => ({
  safe: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val: unknown) => Promise.resolve(val)),
}));

import { useBalanceSheet } from './useBalanceSheet';
import { safe } from '../../shared/lib/api';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useBalanceSheet', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches balance sheet on mount', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      assets: [{ label: 'Cash', amount: 10000, type: 'current' }],
      liabilities: [],
      equity: [],
      totalAssets: 10000,
      totalLiabilities: 0,
      totalEquity: 10000,
    });
    const { result } = renderHook(() => useBalanceSheet(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith(
      '/finance/balance-sheet',
      expect.objectContaining({ params: expect.objectContaining({ as_of: expect.any(String) }) })
    );
  });

  it('defaults assets, liabilities, equity to empty arrays', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue(null);
    const { result } = renderHook(() => useBalanceSheet(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.assets).toEqual([]);
    expect(result.current.data?.liabilities).toEqual([]);
    expect(result.current.data?.equity).toEqual([]);
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useBalanceSheet(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});
