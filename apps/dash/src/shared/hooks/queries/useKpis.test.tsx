import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useKpis, type HubKpis } from './useKpis';

const mockDashboardClient = vi.hoisted(() => ({ getKPIs: vi.fn() }));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../lib/clients', () => ({ dashboardClient: mockDashboardClient }));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

const mockKpis: HubKpis = {
  revenueToday: 5200,
  revenueTrendPct: 12.5,
  ordersToday: 48,
  ordersTrendPct: -3.2,
  headcount: 15,
  activeStaff: 12,
  lowStockItems: 3,
  openInvoices: 7,
  newCustomers: 9,
  apiHealthPct: 99.5,
  revenueSeries: [100, 200, 300],
  ordersSeries: [10, 20, 30],
};

describe('useKpis', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches KPI data successfully', async () => {
    mockDashboardClient.getKPIs.mockResolvedValue(mockKpis);

    const { result } = renderHook(() => useKpis(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.revenueToday).toBe(5200);
    expect(result.current.data?.ordersToday).toBe(48);
    expect(result.current.data?.revenueSeries).toEqual([100, 200, 300]);
  });

  it('calls the correct endpoint with range parameter', async () => {
    mockDashboardClient.getKPIs.mockResolvedValue(mockKpis);

    renderHook(() => useKpis('7d'), { wrapper: createWrapper() });
    await waitFor(() => expect(mockDashboardClient.getKPIs).toHaveBeenCalledWith({ range: '7d' }));
  });

  it('handles API error', async () => {
    mockDashboardClient.getKPIs.mockRejectedValue(new Error('Server error'));

    const { result } = renderHook(() => useKpis(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('returns all expected fields', async () => {
    mockDashboardClient.getKPIs.mockResolvedValue(mockKpis);

    const { result } = renderHook(() => useKpis(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    const data = result.current.data!;
    expect(data).toHaveProperty('revenueToday');
    expect(data).toHaveProperty('revenueTrendPct');
    expect(data).toHaveProperty('headcount');
    expect(data).toHaveProperty('activeStaff');
    expect(data).toHaveProperty('lowStockItems');
    expect(data).toHaveProperty('apiHealthPct');
  });

  it('has a 30s stale time', async () => {
    mockDashboardClient.getKPIs.mockResolvedValue(mockKpis);

    const wrapper = createWrapper();
    const { result } = renderHook(() => useKpis(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    // Re-render the same hook with same wrapper should not refetch (stale time)
    const { result: result2 } = renderHook(() => useKpis(), { wrapper });
    await waitFor(() => expect(result2.current.isSuccess).toBe(true));

    // Only 1 call because of stale time
    expect(mockDashboardClient.getKPIs).toHaveBeenCalledTimes(1);
  });
});
