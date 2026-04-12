import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/financeApi', () => ({
  financeApi: {
    getDashboard: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useFinanceDashboard } from './useDashboard';
import { financeApi } from '../../shared/services/financeApi';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useFinanceDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches finance dashboard data', async () => {
    const data = {
      totalRevenueMTD: 50000,
      totalExpensesMTD: 20000,
      netProfit: 30000,
      outstandingInvoicesTotal: 5000,
      outstandingInvoicesCount: 3,
      totalRevenueMTDChange: '+10%',
      totalExpensesMTDChange: '-5%',
      netProfitChange: '+15%',
      recentTransactions: [],
      overdueInvoices: [],
      barData: [],
    };
    (financeApi.getDashboard as any).mockResolvedValue(data);
    const { result } = renderHook(() => useFinanceDashboard(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (financeApi.getDashboard as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useFinanceDashboard(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });

  it('handles error state', async () => {
    (financeApi.getDashboard as any).mockRejectedValue(new Error('Network error'));
    const { result } = renderHook(() => useFinanceDashboard(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });
});
