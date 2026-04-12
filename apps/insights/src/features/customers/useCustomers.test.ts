import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getCustomerAnalytics: vi.fn(),
    getTopCustomers: vi.fn(),
    getCustomerChurn: vi.fn(),
    getCustomerLtv: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useCustomerAnalytics,
  useTopCustomers,
  useCustomerChurn,
  useCustomerLtv,
} from './useCustomers';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useCustomerAnalytics', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches customer analytics on mount', async () => {
    const data = { totalCustomers: 300, newCustomers: 25, returningCustomers: 275 };
    (insightsService.getCustomerAnalytics as any).mockResolvedValue(data);
    const { result } = renderHook(() => useCustomerAnalytics(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (insightsService.getCustomerAnalytics as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useCustomerAnalytics(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useTopCustomers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches top customers', async () => {
    const input = [{ id: 1, name: 'VIP Corp', totalSpend: 50000 }];
    (insightsService.getTopCustomers as any).mockResolvedValue(input);
    const { result } = renderHook(() => useTopCustomers(10), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([{ id: '1', name: 'VIP Corp', totalSpend: 50000 }]);
    expect(insightsService.getTopCustomers).toHaveBeenCalledWith({ limit: 10 });
  });
});

describe('useCustomerChurn', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches customer churn data', async () => {
    const data = { churnRate: 0.05 };
    (insightsService.getCustomerChurn as any).mockResolvedValue(data);
    const { result } = renderHook(() => useCustomerChurn(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});

describe('useCustomerLtv', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches customer LTV data', async () => {
    const data = { average: 1200, avgLtv: 1200, medianLtv: 900 };
    (insightsService.getCustomerLtv as any).mockResolvedValue(data);
    const { result } = renderHook(() => useCustomerLtv(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});
