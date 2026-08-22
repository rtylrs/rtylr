import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';
import type {
  CreateInvoiceRequest,
  InvoiceSummary,
  InvoiceDetail,
} from '@voxire/voxie/finance/invoices/v1/types';
import { InvoicePaymentServiceClient } from '@voxire/voxie/finance/invoices/payments/v1/client';
import type { Invoice } from '../../shared/services/financeApi';
import { invoiceSchema } from '../../shared/schemas/invoices';
import { useListInvoices } from '../../shared/hooks/queries/useInvoices';

const invoicesClient = createOneKitClient(InvoicesServiceClient, 'finance');
const paymentsClient = createOneKitClient(InvoicePaymentServiceClient, 'finance');

// finance.invoices.v1.Invoice statuses are lowercase (draft/sent/viewed/
// partial/paid/overdue/cancelled/void); the UI works in capitalized labels
// with only 4 states, so extra backend states fold into the closest one.
const STATUS_FROM_API: Record<string, Invoice['status']> = {
  draft: 'Draft',
  sent: 'Sent',
  viewed: 'Sent',
  partial: 'Sent',
  paid: 'Paid',
  overdue: 'Overdue',
  cancelled: 'Cancelled',
  void: 'Cancelled',
};

function fromApiSummary(inv: InvoiceSummary): Invoice {
  return invoiceSchema.parse({
    id: inv.id,
    client: inv.clientName,
    amount: inv.total,
    currency: inv.currency,
    paid: inv.amountPaid,
    status: STATUS_FROM_API[inv.status ?? ''] ?? 'Draft',
    dueDate: inv.dueDate,
    issuedDate: inv.issueDate,
    number: inv.invoiceNumber,
  }) as Invoice;
}

function fromApiDetail(inv: InvoiceDetail): Invoice {
  return invoiceSchema.parse({
    id: inv.id,
    client: inv.clientName,
    clientEmail: inv.clientEmail,
    amount: inv.total,
    currency: inv.currency,
    paid: inv.amountPaid,
    status: STATUS_FROM_API[inv.status ?? ''] ?? 'Draft',
    dueDate: inv.dueDate,
    issuedDate: inv.issueDate,
    items: (inv.items ?? []).map((item) => ({
      description: item.description,
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      amount: item.total,
    })),
    notes: inv.notes,
    discount: inv.discountValue,
    tax: inv.taxAmount,
    subtotal: inv.subtotal,
    number: inv.invoiceNumber,
  }) as Invoice;
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
  const query = useListInvoices({ page, pageSize });
  return {
    ...query,
    data: query.data
      ? {
          invoices: z
            .array(invoiceSchema)
            .parse((query.data.invoices ?? []).map(fromApiSummary)) as Invoice[],
          hasMore: query.data.hasMore ?? false,
          page: query.data.page ?? page,
          pageSize: query.data.pageSize ?? pageSize,
        }
      : undefined,
  };
}

export function useInvoiceDetail(id: string | undefined) {
  return useQuery<Invoice>({
    queryKey: ['invoice', id],
    queryFn: async (): Promise<Invoice> => {
      const res = await invoicesClient.getInvoice({ id: id ?? '' });
      if (!res.invoice) throw new Error('Invoice not found');
      return fromApiDetail(res.invoice);
    },
    enabled: !!id,
  });
}

export function useCreateInvoice() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (data: CreateInvoiceRequest): Promise<Invoice> => {
      const res = await invoicesClient.createInvoice(data);
      if (!res.invoice) throw new Error('Invoice created but not returned');
      return fromApiDetail(res.invoice);
    },
    onSuccess: (created) => {
      qc.invalidateQueries({ queryKey: ['invoices'] });
      if (created?.id) qc.invalidateQueries({ queryKey: ['invoice', created.id] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

export function useUpdateInvoice() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async ({
      id,
      data,
    }: {
      id: string;
      data: Omit<CreateInvoiceRequest, never>;
    }): Promise<Invoice> => {
      const res = await invoicesClient.updateInvoice({ id, body: data });
      if (!res.invoice) throw new Error('Invoice updated but not returned');
      return fromApiDetail(res.invoice);
    },
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['invoices'] });
      qc.invalidateQueries({ queryKey: ['invoice', variables.id] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

export function useDeleteInvoice() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (id: string) => invoicesClient.deleteInvoice({ id }),
    onSuccess: (_data, id) => {
      qc.invalidateQueries({ queryKey: ['invoices'] });
      qc.removeQueries({ queryKey: ['invoice', id] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

const STATUS_TO_API: Record<Invoice['status'], string> = {
  Draft: 'draft',
  Sent: 'sent',
  Paid: 'paid',
  Overdue: 'overdue',
  Cancelled: 'cancelled',
};

export function useUpdateInvoiceStatus() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Invoice['status'] }) =>
      invoicesClient.updateInvoiceStatus({ id, body: { status: STATUS_TO_API[status] } }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['invoices'] });
      qc.invalidateQueries({ queryKey: ['invoice', variables.id] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

export function useRecordPayment() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
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
    }) =>
      paymentsClient.recordPayment({
        invoiceId: id,
        amount,
        currency,
        paymentDate,
        method: method ?? 'cash',
        reference,
        notes: '',
      }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['invoices'] });
      qc.invalidateQueries({ queryKey: ['invoice', variables.id] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

export interface BulkSendResult {
  succeeded: string[];
  failed: string[];
}

// Bulk send loops sendInvoice per id and aggregates per-id outcomes; callers
// toast an honest summary and clear only the succeeded selections.
export async function sendInvoicesByIds(ids: string[]): Promise<BulkSendResult> {
  const results = await Promise.allSettled(ids.map((id) => invoicesClient.sendInvoice({ id })));
  return {
    succeeded: ids.filter((_, i) => results[i].status === 'fulfilled'),
    failed: ids.filter((_, i) => results[i].status === 'rejected'),
  };
}

export interface InvoiceExportResult {
  invoices: InvoiceSummary[];
  pagesFetched: number;
  truncated: boolean;
}

export const MAX_EXPORT_PAGES = 20;

// CSV exports must cover every invoice, not just the current page. Pages are
// looped until hasMore is false or MAX_EXPORT_PAGES is hit (truncated: true so
// the caller can flag the partial export).
export async function fetchAllInvoicesForExport(pageSize = 100): Promise<InvoiceExportResult> {
  const invoices: InvoiceSummary[] = [];
  let pagesFetched = 0;
  let truncated = false;
  while (pagesFetched < MAX_EXPORT_PAGES) {
    pagesFetched += 1;
    const res = await invoicesClient.listInvoices({ page: pagesFetched, pageSize });
    for (const inv of res.invoices ?? []) invoices.push(inv);
    if (!res.hasMore) return { invoices, pagesFetched, truncated: false };
  }
  truncated = true;
  return { invoices, pagesFetched, truncated };
}
