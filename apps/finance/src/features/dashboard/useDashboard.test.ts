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

import { useFinanceDashboard } from './useDashboard';
import { safe } from '../../shared/lib/api';

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
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue(data);
    const { result } = renderHook(() => useFinanceDashboard(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/dashboard');
    expect(result.current.data?.netProfitMtd).toBe(30000);
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useFinanceDashboard(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('handles error state', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Network error'));
    const { result } = renderHook(() => useFinanceDashboard(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });
});
