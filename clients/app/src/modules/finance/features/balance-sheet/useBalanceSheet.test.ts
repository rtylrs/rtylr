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

  it('asks for the chosen date and keeps the ledger lines', async () => {
    mockClient.getBalanceSheet.mockResolvedValue({
      asOf: '2026-10-31',
      lines: [{ accountId: '7', code: '1010', name: 'Bank', type: 'asset', amount: 465 }],
      netIncome: 465,
      totalAssets: 465,
      totalLiabilities: 0,
      totalEquity: 465,
      cashBalance: 8000,
      accountsReceivable: 2000,
      accountsPayable: 0,
    });
    const { result } = renderHook(() => useBalanceSheet('2026-10-31'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.getBalanceSheet).toHaveBeenCalledWith({ asOf: '2026-10-31' });
    expect(result.current.data?.lines).toEqual([
      { accountId: '7', code: '1010', name: 'Bank', type: 'asset', amount: 465 },
    ]);
    expect(result.current.data?.totalEquity).toBe(465);
  });

  it('starts in loading state', () => {
    mockClient.getBalanceSheet.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useBalanceSheet('2026-10-31'), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});
