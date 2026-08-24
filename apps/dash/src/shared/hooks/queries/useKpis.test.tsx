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

// hub.dashboard.v1.GetKPIsResponse. The previous fixture was twelve fields that
// the contract does not carry, so it agreed with the schema and disagreed with
// the server, which is how rtylr#947 survived a green suite.
const mockKpis: HubKpis = {
  revenue: 5200,
  orderCount: 48,
  staffCount: 15,
  customerCount: 9,
  lowStockCount: 3,
  activeApps: 4,
  openInvoices: 0,
};

describe('useKpis', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches KPI data successfully', async () => {
    mockDashboardClient.getKPIs.mockResolvedValue(mockKpis);

    const { result } = renderHook(() => useKpis(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.revenue).toBe(5200);
    expect(result.current.data?.orderCount).toBe(48);
    expect(result.current.data?.customerCount).toBe(9);
  });

  it('sends no range, because the request message has no such field', async () => {
    // This test used to assert getKPIs was called with { range: '7d' }, which
    // locked in the defect: GetKPIsRequest is an empty message, `{ range }`
    // compiled only because excess-property checks do not apply to one, and the
    // generated encoder dropped it. The assertion described the call the client
    // made rather than the call the contract accepts (rtylr#946).
    mockDashboardClient.getKPIs.mockResolvedValue(mockKpis);

    renderHook(() => useKpis(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockDashboardClient.getKPIs).toHaveBeenCalledWith({}));
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
    expect(data).toHaveProperty('revenue');
    expect(data).toHaveProperty('orderCount');
    expect(data).toHaveProperty('staffCount');
    expect(data).toHaveProperty('customerCount');
    expect(data).toHaveProperty('lowStockCount');
    expect(data).toHaveProperty('activeApps');
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
