import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';
import type {
  CreateInvoiceRequest,
  UpdateInvoiceBody,
  InvoiceSummary,
  InvoiceDetail,
  Payment,
} from '@voxire/voxie/finance/invoices/v1/types';
import { InvoicePaymentServiceClient } from '@voxire/voxie/finance/invoices/payments/v1/client';
import type { Invoice } from '../../shared/services/financeApi';
import { invoiceSchema } from '../../shared/schemas/invoices';
import {
  useListInvoices,
  withIdempotencyKey,
  type Keyed,
} from '../../shared/hooks/queries/useInvoices';
import type { RecordPaymentResponse } from '@voxire/voxie/finance/invoices/payments/v1/types';

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

function fromApiDetail(inv: InvoiceDetail, payments: Payment[] = []): Invoice {
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
      unit: item.unit,
      unitPrice: item.unitPrice,
      taxRate: item.taxRate,
      discountPct: item.discountPct,
      sortOrder: item.sortOrder,
      amount: item.total,
    })),
    notes: inv.notes,
    discount: inv.discountValue,
    discountType: inv.discountType,
    discountAmount: inv.discountAmount,
    tax: inv.taxAmount,
    subtotal: inv.subtotal,
    number: inv.invoiceNumber,
    history: payments.map((payment) => ({
      id: payment.id,
      label: [payment.method, payment.reference].filter(Boolean).join(' · ') || 'Payment',
      at: payment.paymentDate ?? '',
    })),
  }) as Invoice;
}

interface UseInvoicesParams {
  page?: number;
  pageSize?: number;
  status?: string;
}

export function useInvoices(params: UseInvoicesParams = {}) {
  const { page = 1, pageSize = 50, status } = params;
  const query = useListInvoices({ page, pageSize, status });
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
  const { t } = useTranslation('financeCommon');
  return useQuery<Invoice>({
    queryKey: ['finance', 'invoice', id],
    queryFn: async (): Promise<Invoice> => {
      const res = await invoicesClient.getInvoice({ id: id ?? '' });
      if (!res.invoice) throw new Error(t('financeInvoices.notFound'));
      return fromApiDetail(res.invoice, res.payments ?? []);
    },
    enabled: !!id,
  });
}

export function useCreateInvoice() {
  const qc = useQueryClient();
  const { t } = useTranslation('financeCommon');
  return useMutation({
    mutationFn: async (data: CreateInvoiceRequest): Promise<Invoice> => {
      const res = await invoicesClient.createInvoice(data);
      if (!res.invoice) throw new Error(t('financeInvoices.createdNotReturned'));
      return fromApiDetail(res.invoice);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['finance', 'invoices'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('financeErrors.generic'))),
  });
}

export function useUpdateInvoice() {
  const qc = useQueryClient();
  const { t } = useTranslation('financeCommon');
  return useMutation({
    mutationFn: async ({
      id,
      data,
    }: {
      id: string;
      // UpdateInvoiceBody, which is what the client's `body` parameter takes.
      //
      // This was `Omit<CreateInvoiceRequest, never>`, a no-op Omit that made the
      // create shape structurally assignable to the update shape. It stopped
      // being assignable when client_name was relaxed from @required to
      // optional: clientName became `string | undefined` on the create side
      // while the update body still requires it. Naming the real type is both
      // the fix and the more honest signature, since this mutation only ever
      // calls updateInvoice.
      data: UpdateInvoiceBody;
    }): Promise<Invoice> => {
      const res = await invoicesClient.updateInvoice({ id, body: data });
      if (!res.invoice) throw new Error(t('financeInvoices.updatedNotReturned'));
      return fromApiDetail(res.invoice);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['finance', 'invoices'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('financeErrors.generic'))),
  });
}

export function useDeleteInvoice() {
  const qc = useQueryClient();
  const { t } = useTranslation('financeCommon');
  return useMutation({
    mutationFn: (id: string) => invoicesClient.deleteInvoice({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['finance', 'invoices'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('financeErrors.generic'))),
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
  const { t } = useTranslation('financeCommon');
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Invoice['status'] }) =>
      invoicesClient.updateInvoiceStatus({ id, body: { status: STATUS_TO_API[status] } }),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['finance', 'invoices'] });
      qc.invalidateQueries({ queryKey: ['finance', 'invoice', variables.id] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('financeErrors.generic'))),
  });
}

interface RecordPaymentVariables {
  id: string;
  amount: number;
  currency: string;
  paymentDate: string;
  method?: string;
  reference: string;
}

export function useRecordPayment() {
  const qc = useQueryClient();
  const { t } = useTranslation('financeCommon');
  const mutation = useMutation<RecordPaymentResponse, Error, Keyed<RecordPaymentVariables>>({
    mutationFn: ({ idempotencyKey, id, amount, currency, paymentDate, method, reference }) =>
      // The generated client already accepts per-call RequestOptions, so the
      // key reaches the wire without a contract change. Without it the server
      // answers this POST with 428 and the payment is never recorded.
      paymentsClient.recordPayment(
        {
          invoiceId: id,
          amount,
          currency,
          paymentDate,
          method: method ?? 'cash',
          reference,
          notes: '',
        },
        { headers: { 'Idempotency-Key': idempotencyKey } }
      ),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['finance', 'invoices'] });
      qc.invalidateQueries({ queryKey: ['finance', 'invoice', variables.id] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('financeErrors.generic'))),
  });

  return withIdempotencyKey(mutation);
}
