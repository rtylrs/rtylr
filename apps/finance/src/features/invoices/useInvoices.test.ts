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

import {
  useInvoices,
  useInvoiceDetail,
  useCreateInvoice,
  useUpdateInvoice,
  useDeleteInvoice,
  useUpdateInvoiceStatus,
  useRecordPayment,
} from './useInvoices';
import { safe } from '../../shared/lib/api';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useInvoices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches invoices on mount', async () => {
    const invoices = [
      {
        id: 'inv-1',
        amount: 500,
        client: 'Acme',
        dueDate: '2026-05-01',
        issuedDate: '2026-04-01',
        status: 'Draft' as const,
      },
    ];
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue(invoices);
    const { result } = renderHook(() => useInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/invoices');
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useInvoices(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useInvoiceDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches invoice detail by id', async () => {
    const inv = {
      id: 'inv-5',
      amount: 1200,
      client: 'Beta Corp',
      dueDate: '2026-06-01',
      issuedDate: '2026-05-01',
      status: 'Sent' as const,
    };
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue(inv);
    const { result } = renderHook(() => useInvoiceDetail('inv-5'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/invoices/inv-5');
  });

  it('does not fetch when id is undefined', () => {
    const { result } = renderHook(() => useInvoiceDetail(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
    expect(safe.get).not.toHaveBeenCalled();
  });
});

describe('useCreateInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls POST /finance/invoices', async () => {
    (safe.post as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'inv-new' });
    const { result } = renderHook(() => useCreateInvoice(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ amount: 300 } as any));
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices', { amount: 300 });
  });
});

describe('useUpdateInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls PUT /finance/invoices/:id', async () => {
    (safe.put as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 'inv-1' });
    const { result } = renderHook(() => useUpdateInvoice(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: 'inv-1', data: { amount: 600 } as any }));
    expect(safe.put).toHaveBeenCalledWith('/finance/invoices/inv-1', { amount: 600 });
  });
});

describe('useDeleteInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls DELETE /finance/invoices/:id', async () => {
    (safe.delete as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    const { result } = renderHook(() => useDeleteInvoice(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync('inv-3'));
    expect(safe.delete).toHaveBeenCalledWith('/finance/invoices/inv-3');
  });
});

describe('useUpdateInvoiceStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls PUT /finance/invoices/:id/status', async () => {
    (safe.put as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    const { result } = renderHook(() => useUpdateInvoiceStatus(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: 'inv-1', status: 'Paid' as any }));
    expect(safe.put).toHaveBeenCalledWith('/finance/invoices/inv-1/status', { status: 'Paid' });
  });
});

describe('useRecordPayment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls POST /finance/invoices/payment', async () => {
    (safe.post as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    const { result } = renderHook(() => useRecordPayment(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: '42', amount: 250 }));
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices/payment', {
      invoiceId: 42,
      amount: 250,
      paymentDate: undefined,
      method: undefined,
    });
  });
});
