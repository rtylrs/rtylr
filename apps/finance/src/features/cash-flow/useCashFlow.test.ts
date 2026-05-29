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

import { useCashFlow } from './useCashFlow';
import { safe } from '../../shared/lib/api';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useCashFlow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches cash flow data on mount', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      points: [
        { month: 'Jan', inflow: 5000, outflow: 3000 },
        { month: 'Feb', inflow: 6000, outflow: 4000 },
      ],
    });
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith(
      '/finance/reports/cash-flow',
      expect.objectContaining({ params: expect.objectContaining({ from: expect.any(String) }) })
    );
    expect(result.current.data).toHaveLength(2);
  });

  it('returns empty array when no data', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ points: [] });
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('parses inflow and outflow from data points', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      points: [{ month: 'Mar', inflow: 8000, outflow: 5000 }],
    });
    const { result } = renderHook(() => useCashFlow(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.inflow).toBe(8000);
    expect(result.current.data?.[0]?.outflow).toBe(5000);
  });
});
