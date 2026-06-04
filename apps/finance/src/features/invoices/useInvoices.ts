import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import type { RecordPaymentResponse } from '@voxire/contracts/finance/invoices/payments/v1/record_payment';
import { safe } from '../../shared/lib/api';
import type { Invoice } from '../../shared/services/financeApi';
import { invoiceSchema } from '../../shared/schemas/invoices';
import { unwrapList } from '../../shared/lib/unwrapList';

interface UseInvoicesParams {
  page?: number;
  pageSize?: number;
}

export interface InvoiceListResult {
  invoices: Invoice[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export function useInvoices(params: UseInvoicesParams = {}) {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<InvoiceListResult>({
    queryKey: ['invoices', { page, pageSize }],
    queryFn: async (): Promise<InvoiceListResult> => {
      const res = await unwrap(
        safe.get<unknown>('/finance/invoices', { params: { page, page_size: pageSize } })
      );
      const raw = res as Record<string, unknown>;
      const invoices = z.array(invoiceSchema).parse(unwrapList(res, 'invoices')) as Invoice[];
      return {
        invoices,
        hasMore: Boolean(raw.has_more),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.page_size ?? pageSize),
      };
    },
  });
}

export function useInvoiceDetail(id: string | undefined) {
  return useQuery<Invoice>({
    queryKey: ['invoice', id],
    queryFn: async (): Promise<Invoice> => {
      // Backend GetInvoiceHandler is registered at /finance/invoices/detail and
      // reads the invoice id from the request (not a path param).
      const res = await unwrap(
        safe.get<unknown>('/finance/invoices/detail', { params: { id: Number(id) } })
      );
      return invoiceSchema.parse(res) as Invoice;
    },
    enabled: !!id,
  });
}

export function useCreateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Invoice>): Promise<Invoice> =>
      unwrap(safe.post<Invoice>('/finance/invoices', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to create invoice'),
  });
}

export function useUpdateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    // Backend UpdateInvoiceHandler is registered at PUT /finance/invoices and
    // reads the id from the request body.
    mutationFn: ({ id, data }: { id: string; data: Partial<Invoice> }): Promise<Invoice> =>
      unwrap(safe.put<Invoice>('/finance/invoices', { id: Number(id), ...data })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to update invoice'),
  });
}

export function useDeleteInvoice() {
  const qc = useQueryClient();
  return useMutation({
    // Backend DeleteInvoiceHandler is registered at DELETE /finance/invoices and
    // reads the id from the request body.
    mutationFn: (id: string): Promise<void> =>
      unwrap(safe.delete<void>('/finance/invoices', { data: { id: Number(id) } })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to delete invoice'),
  });
}

export function useUpdateInvoiceStatus() {
  const qc = useQueryClient();
  return useMutation({
    // Backend UpdateInvoiceStatusHandler is registered at PUT /finance/invoices/status
    // and reads { id, status } from the request body.
    mutationFn: ({ id, status }: { id: string; status: Invoice['status'] }): Promise<void> =>
      unwrap(safe.put<void>('/finance/invoices/status', { id: Number(id), status })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: () => toast.error('Failed to update invoice status'),
  });
}

export function useRecordPayment() {
  const qc = useQueryClient();
  return useMutation({
    // Backend RecordPaymentHandler returns { payment, message }; expose the typed
    // response so callers can read the recorded payment id/details.
    mutationFn: ({
      id,
      amount,
      paymentDate,
      method,
    }: {
      id: string;
      amount: number;
      paymentDate?: string;
      method?: string;
    }): Promise<RecordPaymentResponse> =>
      unwrap(
        safe.post<RecordPaymentResponse>('/finance/invoices/payment', {
          invoiceId: Number(id),
          amount,
          paymentDate,
          method,
        })
      ),
    onSuccess: (
      _data: RecordPaymentResponse,
      variables: { id: string; amount: number; paymentDate?: string; method?: string }
    ) => {
      qc.invalidateQueries({ queryKey: ['invoices'] });
      qc.invalidateQueries({ queryKey: ['invoice', variables.id] });
    },
    onError: () => toast.error('Failed to record payment'),
  });
}
