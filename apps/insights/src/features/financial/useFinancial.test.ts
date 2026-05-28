import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getProfitLoss: vi.fn(),
    getExpenseBreakdown: vi.fn(),
    getInvoiceAging: vi.fn(),
    getCashFlow: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useProfitLoss, useExpenseBreakdown, useInvoiceAging, useCashFlow } from './useFinancial';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useProfitLoss', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches profit and loss data on mount', async () => {
    const data = { revenue: 50000, expenses: 30000, netProfit: 20000 };
    (insightsService.getProfitLoss as any).mockResolvedValue(data);
    const { result } = renderHook(() => useProfitLoss(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes date params to the service', async () => {
    (insightsService.getProfitLoss as any).mockResolvedValue({});
    const params = { start: '2026-01-01', end: '2026-01-31' };
    const { result } = renderHook(() => useProfitLoss(params), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(insightsService.getProfitLoss).toHaveBeenCalledWith(params);
  });

  it('starts in loading state', () => {
    (insightsService.getProfitLoss as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useProfitLoss(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useExpenseBreakdown', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches expense breakdown', async () => {
    const data = { category: 'rent', amount: 5000 };
    (insightsService.getExpenseBreakdown as any).mockResolvedValue(data);
    const { result } = renderHook(() => useExpenseBreakdown(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes date range params to service', async () => {
    (insightsService.getExpenseBreakdown as any).mockResolvedValue({});
    const params = { start: '2026-01-01', end: '2026-03-31' };
    const { result } = renderHook(() => useExpenseBreakdown(params), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(insightsService.getExpenseBreakdown).toHaveBeenCalledWith(params);
  });
});

describe('useInvoiceAging', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches invoice aging data', async () => {
    const data = { current: 10000, days30: 5000, days60: 2000 };
    (insightsService.getInvoiceAging as any).mockResolvedValue(data);
    const { result } = renderHook(() => useInvoiceAging(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});

describe('useCashFlow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches cash flow data', async () => {
    const data = { inflow: 80000, outflow: 60000, net: 20000 };
    (insightsService.getCashFlow as any).mockResolvedValue(data);
    const { result } = renderHook(() => useCashFlow('monthly'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(insightsService.getCashFlow).toHaveBeenCalledWith({ period: 'monthly' });
  });

  it('passes undefined period when not specified', async () => {
    (insightsService.getCashFlow as any).mockResolvedValue({});
    const { result } = renderHook(() => useCashFlow(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(insightsService.getCashFlow).toHaveBeenCalledWith({ period: undefined });
  });
});
