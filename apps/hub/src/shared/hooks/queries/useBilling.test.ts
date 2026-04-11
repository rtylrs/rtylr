import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { useBilling, useInvoices } from './useBilling';

const mockGet = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../lib/api', () => ({
  safe: { get: (...args: any[]) => mockGet(...args) },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useBilling', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches billing data and maps plan limits (pro)', async () => {
    mockGet.mockResolvedValue({
      plan: 'pro',
      status: 'active',
      nextBillingDate: '2026-05-01',
      currentUsage: 8,
    });

    const { result } = renderHook(() => useBilling(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.plan).toBe('pro');
    expect(result.current.data?.amount).toBe(4.99);
    expect(result.current.data?.usage.seats.limit).toBe(20);
    expect(result.current.data?.usage.seats.used).toBe(8);
    expect(result.current.data?.renewsAt).toBe('2026-05-01');
  });

  it('defaults to free plan when plan is missing', async () => {
    mockGet.mockResolvedValue({});

    const { result } = renderHook(() => useBilling(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.plan).toBe('free');
    expect(result.current.data?.amount).toBe(0);
    expect(result.current.data?.usage.seats.limit).toBe(5);
    expect(result.current.data?.usage.orders.limit).toBe(500);
  });

  it('uses scale plan limits', async () => {
    mockGet.mockResolvedValue({ plan: 'scale', currentUsage: 50 });

    const { result } = renderHook(() => useBilling(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.usage.seats.limit).toBe(999);
    expect(result.current.data?.amount).toBe(2.99);
  });

  it('defaults status to active', async () => {
    mockGet.mockResolvedValue({ plan: 'free' });

    const { result } = renderHook(() => useBilling(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.status).toBe('active');
  });

  it('handles API error', async () => {
    mockGet.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useBilling(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('sets currency to USD', async () => {
    mockGet.mockResolvedValue({ plan: 'pro' });

    const { result } = renderHook(() => useBilling(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.currency).toBe('USD');
  });
});

describe('useInvoices', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches invoices from .invoices key', async () => {
    mockGet.mockResolvedValue({
      invoices: [
        {
          id: 'inv_1',
          number: 'INV-001',
          amount: 49.99,
          currency: 'USD',
          status: 'paid',
          createdAt: '2026-01-01',
        },
      ],
    });

    const { result } = renderHook(() => useInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(1);
    expect(result.current.data![0].number).toBe('INV-001');
  });

  it('returns empty array for null response', async () => {
    mockGet.mockResolvedValue(null);

    const { result } = renderHook(() => useInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([]);
  });
});
