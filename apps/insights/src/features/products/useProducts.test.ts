import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getProductPerformance: vi.fn(),
    getBestSellers: vi.fn(),
    getSlowMovers: vi.fn(),
    getMarginAnalysis: vi.fn(),
    getCategoryBreakdown: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useProductPerformance,
  useBestSellers,
  useSlowMovers,
  useMarginAnalysis,
  useCategoryBreakdown,
} from './useProducts';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useProductPerformance', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches product performance on mount', async () => {
    const data = { totalSkus: 120, activeSkus: 110, outOfStock: 4, avgMarginPct: 32 };
    (insightsService.getProductPerformance as any).mockResolvedValue(data);
    const { result } = renderHook(() => useProductPerformance(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (insightsService.getProductPerformance as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useProductPerformance(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useBestSellers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches best sellers list', async () => {
    const data = { products: [{ id: '1', name: 'Widget A', revenue: 10000, rank: 1 }] };
    (insightsService.getBestSellers as any).mockResolvedValue(data);
    const { result } = renderHook(() => useBestSellers(5), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(insightsService.getBestSellers).toHaveBeenCalledWith({ limit: 5 });
  });

  it('unwraps the "products" key when present', async () => {
    const data = { products: [{ id: 1, name: 'Widget B' }] };
    (insightsService.getBestSellers as any).mockResolvedValue(data);
    const { result } = renderHook(() => useBestSellers(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(Array.isArray(result.current.data)).toBe(true);
  });
});

describe('useSlowMovers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches slow movers list', async () => {
    const data = { products: [{ id: '2', name: 'Widget C', revenue: 50, rank: 99 }] };
    (insightsService.getSlowMovers as any).mockResolvedValue(data);
    const { result } = renderHook(() => useSlowMovers(10), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(insightsService.getSlowMovers).toHaveBeenCalledWith({ limit: 10 });
  });
});

describe('useMarginAnalysis', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches margin analysis data', async () => {
    const data = { avgMargin: 0.35, totalRevenue: 500000, grossProfit: 175000 };
    (insightsService.getMarginAnalysis as any).mockResolvedValue(data);
    const { result } = renderHook(() => useMarginAnalysis(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});

describe('useCategoryBreakdown', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches category breakdown list', async () => {
    const data = [
      { categoryId: 1, categoryName: 'Food', unitsSold: 600, revenueMinorUnits: 30000 },
    ];
    (insightsService.getCategoryBreakdown as any).mockResolvedValue(data);
    const { result } = renderHook(() => useCategoryBreakdown(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(Array.isArray(result.current.data)).toBe(true);
  });
});
