// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../lib/api', () => ({
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

import { useBalanceSheet, useCashFlow, usePL } from './useReports';
import { safe } from '../../lib/api';

const RANGE = { from: '2026-01-01', to: '2026-06-30' };

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useBalanceSheet', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches balance sheet with date range', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      totalAssets: 15000,
      totalLiabilities: 5000,
      totalEquity: 10000,
      cashBalance: 8000,
      accountsReceivable: 7000,
      accountsPayable: 5000,
    });
    const { result } = renderHook(() => useBalanceSheet(RANGE), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/balance-sheet', { params: RANGE });
    expect(result.current.data?.assets).toHaveLength(2);
  });

  it('does not fetch when range is empty', () => {
    const { result } = renderHook(() => useBalanceSheet({ from: '', to: '' }), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(safe.get).not.toHaveBeenCalled();
  });

  it('starts in loading state when range is valid', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useBalanceSheet(RANGE), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useCashFlow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches cash flow with date range', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      points: [
        { label: 'Jan', inflow: 5000, outflow: 3000 },
        { label: 'Feb', inflow: 6000, outflow: 4000 },
      ],
    });
    const { result } = renderHook(() => useCashFlow(RANGE), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/cash-flow', { params: RANGE });
    expect(result.current.data?.points).toHaveLength(2);
  });

  it('does not fetch when from is empty', () => {
    const { result } = renderHook(() => useCashFlow({ from: '', to: '2026-06-30' }), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('usePL', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches profit and loss with date range', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      period: '2026-Q1',
      revenue: 100000,
      cogs: 40000,
      grossProfit: 60000,
      operatingExpenses: 20000,
      netProfit: 40000,
    });
    const { result } = renderHook(() => usePL(RANGE), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/api/finance/report/v1/pl', {
      params: expect.objectContaining({ period: `${RANGE.from},${RANGE.to}`, date: RANGE.to }),
    });
  });
});
