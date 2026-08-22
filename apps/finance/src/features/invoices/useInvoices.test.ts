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
}));

const mockPaymentsClient = vi.hoisted(() => ({
  recordPayment: vi.fn(),
}));

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
  useRecordPayment,
} from './useInvoices';

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
      expect.objectContaining({ invoiceId: '42', amount: 250 })
    );
  });
});
