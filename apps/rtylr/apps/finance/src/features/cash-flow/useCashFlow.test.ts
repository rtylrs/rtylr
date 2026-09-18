// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('@voxire/api', () => ({
  unwrap: (val: unknown) => Promise.resolve(val),
  oneKitResult: (op: () => Promise<unknown>) => op(),
}));

vi.mock('../../shared/lib/clients', () => ({
  insightsFinancialClient: {
    getCashFlowAnalytics: vi.fn(),
  },
}));

import { useCashFlow } from './useCashFlow';
import { insightsFinancialClient } from '../../shared/lib/clients';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

// These cases previously mocked a `{ series: [{ inflowMinorUnits, ... }] }`
// response. No endpoint has ever returned that shape: the only HTTP surface is
// insights.financial.v1 GetCashFlowAnalytics, which returns
// `{ analytics: { trend: [{ date, inflow, outflow, net }] } }` with the amounts
// already converted to major units. The old suite passed because it asserted
// against the mock rather than against the contract, so it stayed green while
// the chart rendered empty in production (#910).
describe('useCashFlow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const analytics = (trend: unknown[]) => ({ analytics: { trend } });

  it('requests the period and nothing else', async () => {
    vi.mocked(insightsFinancialClient.getCashFlowAnalytics).mockResolvedValue(
      analytics([{ date: '2026-01-01', inflow: 5000, outflow: 3000, net: 2000 }])
    );
    const { result } = renderHook(() => useCashFlow({ period: 'quarter' }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    // The server derives the date range from the period itself, so sending
    // `from`/`to` was always discarded.
    expect(insightsFinancialClient.getCashFlowAnalytics).toHaveBeenCalledWith({
      period: 'quarter',
    });
  });

  it('defaults the period to month', async () => {
    vi.mocked(insightsFinancialClient.getCashFlowAnalytics).mockResolvedValue(analytics([]));
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(insightsFinancialClient.getCashFlowAnalytics).toHaveBeenCalledWith({ period: 'month' });
  });

  it('reads the points from analytics.trend', async () => {
    vi.mocked(insightsFinancialClient.getCashFlowAnalytics).mockResolvedValue(
      analytics([
        { date: '2026-01-01', inflow: 5000, outflow: 3000, net: 2000 },
        { date: '2026-02-01', inflow: 6000, outflow: 4000, net: 2000 },
      ])
    );
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(2);
  });

  it('passes major units through without converting again', async () => {
    vi.mocked(insightsFinancialClient.getCashFlowAnalytics).mockResolvedValue(
      analytics([{ date: '2026-03-01', inflow: 8000, outflow: 5000, net: 3000 }])
    );
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.month).toBe('2026-03-01');
    // 8000, not 80: the handler already applied minorToMajor. Dividing here
    // would have shown every amount at a hundredth of its value.
    expect(result.current.data?.[0]?.inflow).toBe(8000);
    expect(result.current.data?.[0]?.outflow).toBe(5000);
  });

  it('returns an empty array when the trend is empty', async () => {
    vi.mocked(insightsFinancialClient.getCashFlowAnalytics).mockResolvedValue(analytics([]));
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('returns an empty array when analytics is absent entirely', async () => {
    vi.mocked(insightsFinancialClient.getCashFlowAnalytics).mockResolvedValue({});
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('starts in loading state', () => {
    vi.mocked(insightsFinancialClient.getCashFlowAnalytics).mockReturnValue(
      new Promise(() => {}) as never
    );
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});
