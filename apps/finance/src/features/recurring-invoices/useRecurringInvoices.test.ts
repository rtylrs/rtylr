// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
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

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import { useRecurringInvoices, useCreateRecurringInvoice } from './useRecurringInvoices';
import { safe } from '../../shared/lib/api';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useRecurringInvoices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches recurring invoices on mount', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      recurringInvoices: [
        {
          id: 1,
          templateInvoiceId: 42,
          frequency: 'monthly',
          nextInvoiceDate: '2026-07-01',
          isActive: true,
        },
      ],
    });
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/recurring-invoices');
    expect(result.current.data).toHaveLength(1);
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('returns empty array when list is empty', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ recurringInvoices: [] });
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('parses the backend payload shape', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      recurringInvoices: [{ id: 7, templateInvoiceId: 9, frequency: 'weekly', isActive: false }],
    });
    const { result } = renderHook(() => useRecurringInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.id).toBe(7);
    expect(result.current.data?.[0]?.isActive).toBe(false);
  });
});

describe('useCreateRecurringInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls POST /finance/recurring-invoices', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ recurringInvoices: [] });
    (safe.post as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 5 });
    const { result } = renderHook(() => useCreateRecurringInvoice(), { wrapper: createWrapper() });
    await act(() =>
      result.current.mutateAsync({
        templateInvoiceId: 42,
        frequency: 'monthly',
        nextInvoiceDate: '2026-07-01',
        isActive: true,
      })
    );
    expect(safe.post).toHaveBeenCalledWith('/finance/recurring-invoices', expect.any(Object));
  });
});
