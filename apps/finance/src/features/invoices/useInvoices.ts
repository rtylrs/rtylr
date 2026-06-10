import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import type { RecordPaymentResponse } from '@voxire/contracts/finance/invoices/payments/v1/record_payment';
import type { CreateInvoiceRequest } from '@voxire/contracts/finance/invoices/v1/create_invoice';
import type { UpdateInvoiceRequest } from '@voxire/contracts/finance/invoices/v1/update_invoice';
import { validateContractPayload } from '@voxire/contracts/validation/metadata';
import { safe } from '../../shared/lib/api';
import type { Invoice } from '../../shared/services/financeApi';
import { invoiceSchema } from '../../shared/schemas/invoices';
import { unwrapList } from '../../shared/lib/unwrapList';

function assertContractPayload(messageName: string, payload: Record<string, unknown>): void {
  const issues = validateContractPayload(messageName, payload);
  if (issues.length > 0) {
    throw new Error(issues[0].message);
  }
}

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
    mutationFn: (data: CreateInvoiceRequest): Promise<Invoice> => {
      assertContractPayload('finance.invoices.v1.CreateInvoiceRequest', {
        ...(data as unknown as Record<string, unknown>),
      });
      return unwrap(safe.post<Invoice>('/finance/invoices', data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    onError: (error) =>
      toast.error(error instanceof Error ? error.message : 'Failed to create invoice'),
  });
}

export function useUpdateInvoice() {
  const qc = useQueryClient();
  return useMutation({
    // Backend UpdateInvoiceHandler is registered at PUT /finance/invoices and
    // reads the id from the request body.
    mutationFn: ({ id, data }: { id: string; data: Partial<Invoice> }): Promise<Invoice> => {
      const payload = { id: Number(id), ...data } as UpdateInvoiceRequest;
      assertContractPayload('finance.invoices.v1.UpdateInvoiceRequest', {
        ...(payload as unknown as Record<string, unknown>),
      });
      return unwrap(safe.put<Invoice>('/finance/invoices', payload));
    },
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
    mutationFn: ({ id, status }: { id: string; status: Invoice['status'] }): Promise<void> => {
      const payload = { id: Number(id), status };
      assertContractPayload('finance.invoices.v1.UpdateInvoiceStatusRequest', payload);
      return unwrap(safe.put<void>('/finance/invoices/status', payload));
    },
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
      currency,
      paymentDate,
      method,
      reference,
    }: {
      id: string;
      amount: number;
      currency: string;
      paymentDate: string;
      method?: string;
      reference: string;
    }): Promise<RecordPaymentResponse> => {
      const payload = {
        invoiceId: Number(id),
        amount,
        currency,
        paymentDate,
        method,
        reference,
      };
      assertContractPayload('finance.invoices.payments.v1.RecordPaymentRequest', payload);
      return unwrap(safe.post<RecordPaymentResponse>('/finance/invoices/payment', payload));
    },
    onSuccess: (
      _data: RecordPaymentResponse,
      variables: {
        id: string;
        amount: number;
        currency: string;
        paymentDate: string;
        method?: string;
        reference: string;
      }
    ) => {
      qc.invalidateQueries({ queryKey: ['invoices'] });
      qc.invalidateQueries({ queryKey: ['invoice', variables.id] });
    },
    onError: () => toast.error('Failed to record payment'),
  });
}
