// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/lib/clients', () => ({
  financeDashboardClient: {
    getFinanceDashboard: vi.fn(),
  },
}));

import { useFinanceDashboard } from './useDashboard';
import { financeDashboardClient } from '../../shared/lib/clients';

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
    // Mirrors the deployed DashboardResponse (snake_case keys camelCased by
    // the @voxire/api interceptor before the schema runs).
    const data = {
      totalReceivable: 5000,
      overdueAmount: 1200,
      totalExpensesMtd: 20000,
      netProfitMtd: 30000,
    };
    vi.mocked(financeDashboardClient.getFinanceDashboard).mockResolvedValue(data);
    const { result } = renderHook(() => useFinanceDashboard(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(financeDashboardClient.getFinanceDashboard).toHaveBeenCalledWith({});
    expect(result.current.data?.netProfitMtd).toBe(30000);
  });

  it('starts in loading state', () => {
    vi.mocked(financeDashboardClient.getFinanceDashboard).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useFinanceDashboard(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('handles error state', async () => {
    vi.mocked(financeDashboardClient.getFinanceDashboard).mockRejectedValue(
      new Error('Network error')
    );
    const { result } = renderHook(() => useFinanceDashboard(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toMatchObject({ code: 'UNKNOWN' });
  });
});
