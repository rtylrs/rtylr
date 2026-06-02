import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getKpis: vi.fn(),
    getDashboardSummary: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useKpis, useDashboardSummary } from './useDashboard';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useKpis', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches KPIs on mount', async () => {
    const kpis = { revenue: 5000, customerCount: 120 };
    (insightsService.getKpis as any).mockResolvedValue(kpis);
    const { result } = renderHook(() => useKpis(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(kpis);
  });

  it('starts in loading state', () => {
    (insightsService.getKpis as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useKpis(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useDashboardSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches dashboard summary', async () => {
    const summary = {
      revenue: { value: 10000, deltaPct: 0.1 },
      customers: { value: 50, deltaPct: 0.05 },
    };
    (insightsService.getDashboardSummary as any).mockResolvedValue(summary);
    const { result } = renderHook(() => useDashboardSummary(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(summary);
  });

  it('starts in loading state', () => {
    (insightsService.getDashboardSummary as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useDashboardSummary(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});
