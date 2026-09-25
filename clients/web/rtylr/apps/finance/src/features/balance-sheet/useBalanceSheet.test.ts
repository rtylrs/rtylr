// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({
  getBalanceSheet: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
}));

import { useBalanceSheet } from './useBalanceSheet';

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
    mockClient.getBalanceSheet.mockResolvedValue({
      totalAssets: 10000,
      totalLiabilities: 0,
      totalEquity: 10000,
      cashBalance: 8000,
      accountsReceivable: 2000,
      accountsPayable: 0,
    });
    const { result } = renderHook(() => useBalanceSheet(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.getBalanceSheet).toHaveBeenCalledWith({});
    expect(result.current.data?.totalAssets).toBe(10000);
    expect(result.current.data?.assets).toHaveLength(2);
  });

  it('starts in loading state', () => {
    mockClient.getBalanceSheet.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useBalanceSheet(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});
