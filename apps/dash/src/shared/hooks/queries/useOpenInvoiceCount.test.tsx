// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockClient = vi.hoisted(() => ({ getInvoiceAging: vi.fn() }));
vi.mock('../../lib/clients', () => ({ insightsFinancialClient: mockClient }));

import { useOpenInvoiceCount } from './useOpenInvoiceCount';

const wrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

/**
 * rtylr#1001. KpiInvoices read `openInvoices`, which hub.dashboard.v1 does not
 * carry, so it rendered a confident 0 forever.
 *
 * insights.financial.v1 getInvoiceAging already returns per-bucket counts, and
 * dash already calls it for InvoiceAgingWidget. Aging buckets are outstanding
 * invoices by definition (current, 1-30, 31-60, 61-90, 90+), so their counts sum
 * to the number of open invoices.
 */
describe('useOpenInvoiceCount', () => {
  beforeEach(() => vi.clearAllMocks());

  it('sums the counts across every aging bucket', async () => {
    mockClient.getInvoiceAging.mockResolvedValue({
      aging: {
        totalOutstanding: 4200,
        buckets: [
          { bucket: 'current', count: 3, amount: 1000 },
          { bucket: '1-30', count: 2, amount: 2000 },
          { bucket: '90+', count: 1, amount: 1200 },
        ],
      },
    });
    const { result } = renderHook(() => useOpenInvoiceCount(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(6);
  });

  it('returns 0 when nothing is outstanding', async () => {
    mockClient.getInvoiceAging.mockResolvedValue({ aging: { totalOutstanding: 0, buckets: [] } });
    const { result } = renderHook(() => useOpenInvoiceCount(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(0);
  });

  it('tolerates a missing aging object rather than throwing', async () => {
    mockClient.getInvoiceAging.mockResolvedValue({});
    const { result } = renderHook(() => useOpenInvoiceCount(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(0);
  });

  it('coerces an int64-as-string count, which is how the wire carries it', async () => {
    mockClient.getInvoiceAging.mockResolvedValue({
      aging: {
        buckets: [
          { bucket: 'current', count: '4' },
          { bucket: '1-30', count: '2' },
        ],
      },
    });
    const { result } = renderHook(() => useOpenInvoiceCount(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(6);
  });

  it('does not produce NaN when a bucket omits its count', async () => {
    // The discriminating case. An earlier fixture used count: '', where
    // Number('') and toFiniteNumber('') are both 0, so swapping the helper for a
    // raw Number() passed every test. A missing count is what separates them:
    // Number(undefined) is NaN and poisons the whole sum.
    mockClient.getInvoiceAging.mockResolvedValue({
      aging: { buckets: [{ bucket: 'current', count: 3 }, { bucket: '1-30' }] },
    });
    const { result } = renderHook(() => useOpenInvoiceCount(), { wrapper: wrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(3);
  });
});
