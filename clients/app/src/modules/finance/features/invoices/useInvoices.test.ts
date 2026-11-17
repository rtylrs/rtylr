// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import type { UpdateInvoiceBody } from '@voxire/voxie/finance/invoices/v1/types';

const mockInvoicesClient = vi.hoisted(() => ({
  listInvoices: vi.fn(),
  getInvoice: vi.fn(),
  createInvoice: vi.fn(),
  updateInvoice: vi.fn(),
  deleteInvoice: vi.fn(),
  updateInvoiceStatus: vi.fn(),
  setInvoiceProject: vi.fn(),
}));

const mockPaymentsClient = vi.hoisted(() => ({
  recordPayment: vi.fn(),
}));

const mockToast = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }));

vi.mock('sonner', () => ({ toast: mockToast }));

vi.mock('@voxire/api', () => ({
  createOneKitClient: (Client: { name: string }) =>
    Client.name === 'InvoicePaymentServiceClient' ? mockPaymentsClient : mockInvoicesClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

import {
  useInvoices,
  useInvoiceDetail,
  useCreateInvoice,
  useUpdateInvoice,
  useDeleteInvoice,
  useUpdateInvoiceStatus,
  useSetInvoiceProject,
  useRecordPayment,
} from './useInvoices';
import type { Invoice } from '../../shared/services/financeApi';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

const baseDetail = {
  id: '5',
  ownerId: '1',
  invoiceNumber: 'INV-5',
  clientName: 'Beta Corp',
  clientEmail: 'billing@beta.test',
  clientAddress: '',
  clientTaxId: '',
  status: 'sent',
  issueDate: '2026-05-01',
  dueDate: '2026-06-01',
  paidDate: '',
  currency: 'USD',
  subtotal: 1200,
  discountType: 'none',
  discountValue: 0,
  discountAmount: 0,
  taxRate: 0,
  taxAmount: 0,
  total: 1200,
  amountPaid: 0,
  amountDue: 1200,
  notes: '',
  terms: '',
  createdAt: '',
  updatedAt: '',
  items: [],
};

describe('useInvoices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches invoices on mount', async () => {
    mockInvoicesClient.listInvoices.mockResolvedValue({
      invoices: [],
      hasMore: false,
      page: 1,
      pageSize: 50,
    });
    const { result } = renderHook(() => useInvoices(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockInvoicesClient.listInvoices).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1, pageSize: 50 })
    );
  });

  it('starts in loading state', () => {
    mockInvoicesClient.listInvoices.mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useInvoices(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useInvoiceDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches invoice detail by id', async () => {
    mockInvoicesClient.getInvoice.mockResolvedValue({ invoice: baseDetail, payments: [] });
    const { result } = renderHook(() => useInvoiceDetail('5'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockInvoicesClient.getInvoice).toHaveBeenCalledWith({ id: '5' });
    expect(result.current.data?.status).toBe('Sent');
  });

  it('preserves line tax, discount metadata, and payment history from the detail response', async () => {
    mockInvoicesClient.getInvoice.mockResolvedValue({
      invoice: {
        ...baseDetail,
        discountType: 'percentage',
        discountValue: 10,
        discountAmount: 120,
        items: [
          {
            description: 'Taxed line',
            quantity: 1,
            unitPrice: 100,
            taxRate: 10,
            total: 110,
          },
        ],
      },
      payments: [
        {
          id: 'pay-1',
          amount: 50,
          currency: 'USD',
          paymentDate: '2026-06-10',
          method: 'Card',
          reference: 'TXN-1',
        },
      ],
    });
    const { result } = renderHook(() => useInvoiceDetail('5'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.items?.[0]?.taxRate).toBe(10);
    expect(result.current.data?.discountType).toBe('percentage');
    expect(result.current.data?.discountAmount).toBe(120);
    expect(result.current.data?.history).toEqual([
      { id: 'pay-1', label: 'Card · TXN-1', at: '2026-06-10' },
    ]);
  });

  it('maps the tagged Work project and treats a missing project_id as untagged', async () => {
    mockInvoicesClient.getInvoice.mockResolvedValueOnce({
      invoice: { ...baseDetail, projectId: '77' },
      payments: [],
    });
    const tagged = renderHook(() => useInvoiceDetail('5'), { wrapper: createWrapper() });
    await waitFor(() => expect(tagged.result.current.isSuccess).toBe(true));
    expect(tagged.result.current.data?.projectId).toBe('77');

    mockInvoicesClient.getInvoice.mockResolvedValueOnce({ invoice: baseDetail, payments: [] });
    const untagged = renderHook(() => useInvoiceDetail('6'), { wrapper: createWrapper() });
    await waitFor(() => expect(untagged.result.current.isSuccess).toBe(true));
    expect(untagged.result.current.data?.projectId).toBeNull();
  });

  it('does not fetch when id is undefined', () => {
    const { result } = renderHook(() => useInvoiceDetail(undefined), { wrapper: createWrapper() });
    expect(result.current.fetchStatus).toBe('idle');
    expect(mockInvoicesClient.getInvoice).not.toHaveBeenCalled();
  });
});

describe('useCreateInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('creates an invoice', async () => {
    mockInvoicesClient.createInvoice.mockResolvedValue({ invoice: baseDetail });
    const { result } = renderHook(() => useCreateInvoice(), { wrapper: createWrapper() });
    const invoice = {
      clientName: 'Beta Corp',
      clientEmail: 'billing@beta.test',
      issueDate: '2026-05-01',
      dueDate: '2026-06-01',
      currency: 'USD',
      discountType: 'none',
      items: [{ description: 'Consulting', quantity: 1, unitPrice: 300 }],
    };
    await act(() => result.current.mutateAsync(invoice as never));
    expect(mockInvoicesClient.createInvoice).toHaveBeenCalledWith(invoice);
  });
});

describe('useUpdateInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('updates an invoice', async () => {
    mockInvoicesClient.updateInvoice.mockResolvedValue({ invoice: baseDetail });
    const { result } = renderHook(() => useUpdateInvoice(), { wrapper: createWrapper() });
    // A complete UpdateInvoiceBody rather than `as never`. The cast was hiding
    // the fact that this mutation's parameter type did not match the client's
    // body type, which is the defect this test sits next to.
    const invoice: UpdateInvoiceBody = {
      clientName: 'Beta Corp',
      clientEmail: 'beta@example.com',
      issueDate: '2026-01-01',
      dueDate: '2026-01-31',
      currency: 'USD',
      discountType: 'none',
    };
    await act(() => result.current.mutateAsync({ id: '1', data: invoice }));
    expect(mockInvoicesClient.updateInvoice).toHaveBeenCalledWith({ id: '1', body: invoice });
  });
});

describe('useDeleteInvoice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('deletes an invoice', async () => {
    mockInvoicesClient.deleteInvoice.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useDeleteInvoice(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync('3'));
    expect(mockInvoicesClient.deleteInvoice).toHaveBeenCalledWith({ id: '3' });
  });
});

describe('useUpdateInvoiceStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('updates invoice status', async () => {
    mockInvoicesClient.updateInvoiceStatus.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useUpdateInvoiceStatus(), { wrapper: createWrapper() });
    await act(() => result.current.mutateAsync({ id: '1', status: 'Paid' }));
    expect(mockInvoicesClient.updateInvoiceStatus).toHaveBeenCalledWith({
      id: '1',
      body: { status: 'paid' },
    });
  });
});

describe('useSetInvoiceProject', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const withClient = (client: QueryClient) => {
    return ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);
  };

  it('tags a sent invoice and keeps the cached payment history', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const cached = {
      id: '5',
      client: 'Beta Corp',
      amount: 1200,
      status: 'Sent',
      projectId: null,
      history: [{ id: 'p1', label: 'card', at: '2026-05-02' }],
    } as Invoice;
    client.setQueryData(['finance', 'invoice', '5'], cached);
    const invalidate = vi.spyOn(client, 'invalidateQueries');
    mockInvoicesClient.setInvoiceProject.mockResolvedValue({
      invoice: { ...baseDetail, projectId: '7' },
    });

    const { result } = renderHook(() => useSetInvoiceProject(), { wrapper: withClient(client) });
    const updated = await act(() => result.current.mutateAsync({ id: '5', projectId: '7' }));

    expect(mockInvoicesClient.setInvoiceProject).toHaveBeenCalledWith({
      id: '5',
      body: { projectId: '7' },
    });
    expect(updated.projectId).toBe('7');
    expect(client.getQueryData<Invoice>(['finance', 'invoice', '5'])).toEqual({
      ...cached,
      projectId: '7',
    });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['finance', 'invoice', '5'] });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['finance', 'invoices'] });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ['insights-dashboard-project'] });
  });

  it('sends 0 to clear the tag', async () => {
    mockInvoicesClient.setInvoiceProject.mockResolvedValue({ invoice: baseDetail });

    const { result } = renderHook(() => useSetInvoiceProject(), { wrapper: createWrapper() });
    const updated = await act(() => result.current.mutateAsync({ id: '5', projectId: null }));

    expect(mockInvoicesClient.setInvoiceProject).toHaveBeenCalledWith({
      id: '5',
      body: { projectId: '0' },
    });
    expect(updated.projectId).toBeNull();
  });

  it('toasts when the server refuses the change', async () => {
    mockInvoicesClient.setInvoiceProject.mockRejectedValue(new Error('409'));

    const { result } = renderHook(() => useSetInvoiceProject(), { wrapper: createWrapper() });
    await act(async () => {
      await expect(result.current.mutateAsync({ id: '5', projectId: '7' })).rejects.toThrow('409');
    });

    expect(mockToast.error).toHaveBeenCalledWith('financeInvoices.detail.projectChangeFailed');
  });
});

describe('useRecordPayment', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('records a payment', async () => {
    mockPaymentsClient.recordPayment.mockResolvedValue({ message: 'ok' });
    const { result } = renderHook(() => useRecordPayment(), { wrapper: createWrapper() });
    await act(() =>
      result.current.mutateAsync({
        id: '42',
        amount: 250,
        currency: 'USD',
        paymentDate: '2026-06-01',
        method: 'card',
        reference: 'TXN-42',
      })
    );
    expect(mockPaymentsClient.recordPayment).toHaveBeenCalledWith(
      expect.objectContaining({ invoiceId: '42', amount: 250 }),
      { headers: { 'Idempotency-Key': expect.any(String) } }
    );
  });
});
