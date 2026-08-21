// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/lib/clients', () => ({
  financialCompatibilityClient: {
    getCashFlow: vi.fn(),
  },
}));

import { useCashFlow } from './useCashFlow';
import { financialCompatibilityClient } from '../../shared/lib/clients';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useCashFlow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches the cash flow series on mount', async () => {
    vi.mocked(financialCompatibilityClient.getCashFlow).mockResolvedValue({
      series: [
        { date: '2026-01-01', inflowMinorUnits: 500000, outflowMinorUnits: 300000 },
        { date: '2026-02-01', inflowMinorUnits: 600000, outflowMinorUnits: 400000 },
      ],
    });
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(financialCompatibilityClient.getCashFlow).toHaveBeenCalledWith({
      from: expect.any(String),
      to: expect.any(String),
    });
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when no data', async () => {
    vi.mocked(financialCompatibilityClient.getCashFlow).mockResolvedValue({ series: [] });
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('starts in loading state', () => {
    vi.mocked(financialCompatibilityClient.getCashFlow).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('maps minor units to major-unit inflow/outflow points', async () => {
    vi.mocked(financialCompatibilityClient.getCashFlow).mockResolvedValue({
      series: [{ date: '2026-03-01', inflowMinorUnits: 800000, outflowMinorUnits: 500000 }],
    });
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.month).toBe('2026-03-01');
    expect(result.current.data?.[0]?.inflow).toBe(8000);
    expect(result.current.data?.[0]?.outflow).toBe(5000);
  });
});
