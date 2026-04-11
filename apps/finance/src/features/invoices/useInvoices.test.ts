import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/financeApi', () => ({
  financeApi: {
    getInvoices: vi.fn(),
    getInvoiceDetail: vi.fn(),
    createInvoice: vi.fn(),
    updateInvoice: vi.fn(),
    deleteInvoice: vi.fn(),
    updateInvoiceStatus: vi.fn(),
    recordPayment: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
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
import { financeApi } from '../../shared/services/financeApi';

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
    const data = [{ id: 'inv-1', amount: 500 }];
    (financeApi.getInvoices as any).mockResolvedValue(data);
    const { result } = renderHook(() => useInvoices(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (financeApi.getInvoices as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useInvoices(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useInvoiceDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches invoice detail by id', async () => {
    const inv = { id: 'inv-5', amount: 1200 };
    (financeApi.getInvoiceDetail as any).mockResolvedValue(inv);
    const { result } = renderHook(() => useInvoiceDetail('inv-5'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(inv);
  });

  it('does not fetch when id is undefined', () => {
    const { result } = renderHook(() => useInvoiceDetail(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(financeApi.getInvoiceDetail).not.toHaveBeenCalled();
  });
});

describe('useCreateInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createInvoice service method', async () => {
    (financeApi.createInvoice as any).mockResolvedValue({ id: 'inv-new' });
    const { result } = renderHook(() => useCreateInvoice(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ amount: 300 } as any));
    expect(financeApi.createInvoice).toHaveBeenCalledWith({ amount: 300 });
  });
});

describe('useUpdateInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateInvoice with id and data', async () => {
    (financeApi.updateInvoice as any).mockResolvedValue({ id: 'inv-1' });
    const { result } = renderHook(() => useUpdateInvoice(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 'inv-1', data: { amount: 600 } as any }));
    expect(financeApi.updateInvoice).toHaveBeenCalledWith('inv-1', { amount: 600 });
  });
});

describe('useDeleteInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteInvoice with id', async () => {
    (financeApi.deleteInvoice as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useDeleteInvoice(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync('inv-3'));
    expect(financeApi.deleteInvoice).toHaveBeenCalledWith('inv-3');
  });
});

describe('useUpdateInvoiceStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateInvoiceStatus with id and status', async () => {
    (financeApi.updateInvoiceStatus as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useUpdateInvoiceStatus(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 'inv-1', status: 'paid' as any }));
    expect(financeApi.updateInvoiceStatus).toHaveBeenCalledWith('inv-1', 'paid');
  });
});

describe('useRecordPayment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls recordPayment with id and amount', async () => {
    (financeApi.recordPayment as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useRecordPayment(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 'inv-1', amount: 250 }));
    expect(financeApi.recordPayment).toHaveBeenCalledWith('inv-1', 250);
  });
});
