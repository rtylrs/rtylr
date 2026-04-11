import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getRevenue: vi.fn(),
    getRevenueByApp: vi.fn(),
    getRevenueComparison: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useRevenue, useRevenueByApp, useRevenueComparison } from './useRevenue';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useRevenue', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches revenue data on mount', async () => {
    const data = { total: 10000 };
    (insightsService.getRevenue as any).mockResolvedValue(data);
    const { result } = renderHook(() => useRevenue(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('passes period to the service', async () => {
    (insightsService.getRevenue as any).mockResolvedValue({});
    const { result } = renderHook(() => useRevenue('monthly'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(insightsService.getRevenue).toHaveBeenCalledWith({ period: 'monthly' });
  });

  it('starts in loading state', () => {
    (insightsService.getRevenue as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useRevenue(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useRevenueByApp', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches revenue by app', async () => {
    const data = [{ app: 'book', revenue: 3000 }];
    (insightsService.getRevenueByApp as any).mockResolvedValue(data);
    const { result } = renderHook(() => useRevenueByApp('weekly'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(insightsService.getRevenueByApp).toHaveBeenCalledWith({ period: 'weekly' });
  });
});

describe('useRevenueComparison', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches revenue comparison', async () => {
    const data = { current: 5000, previous: 4500 };
    (insightsService.getRevenueComparison as any).mockResolvedValue(data);
    const { result } = renderHook(() => useRevenueComparison('yearly'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(insightsService.getRevenueComparison).toHaveBeenCalledWith({ period: 'yearly' });
  });
});
