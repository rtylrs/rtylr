import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { QuoteServiceClient } from '@voxire/voxie/finance/quote/v1/client';
import type { Quote as ApiQuote } from '@voxire/voxie/finance/quote/v1/types';
import { quoteSchema } from '../../shared/schemas/quotes';

const quoteClient = createOneKitClient(QuoteServiceClient, 'finance');

// finance.quote.v1.Quote.status is lowercase (draft/sent/accepted/rejected/
// expired/converted); the UI works in capitalized labels with "Declined"
// instead of "rejected".
const STATUS_FROM_API: Record<string, string> = {
  draft: 'Draft',
  sent: 'Sent',
  accepted: 'Accepted',
  rejected: 'Declined',
  expired: 'Expired',
  converted: 'Accepted',
};

function fromApiQuote(q: ApiQuote): ReturnType<typeof quoteSchema.parse> {
  return quoteSchema.parse({
    id: q.id,
    client: q.customerName,
    amount: q.total,
    status: STATUS_FROM_API[q.status ?? ''] ?? 'Draft',
    issuedDate: q.createdAt,
    expiresAt: q.validUntil,
  });
}

interface UseQuotesParams {
  page?: number;
  pageSize?: number;
}

export interface QuoteListResult {
  quotes: ReturnType<typeof quoteSchema.parse>[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export function useQuotes(params: UseQuotesParams = {}) {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<QuoteListResult>({
    queryKey: ['quotes', 'filter', { page, pageSize }],
    queryFn: async (): Promise<QuoteListResult> => {
      const res = await quoteClient.listQuotes({ page, pageSize });
      return {
        quotes: (res.quotes ?? []).map(fromApiQuote),
        hasMore: res.hasMore ?? false,
        page: res.page ?? page,
        pageSize: res.pageSize ?? pageSize,
      };
    },
  });
}

interface CreateQuotePayload {
  client: string;
  amount: number;
  currency: string;
  issuedDate: string;
  expiresAt: string;
  description?: string;
}

export function useCreateQuote() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (data: CreateQuotePayload) =>
      quoteClient.createQuote({
        customerName: data.client,
        customerEmail: '',
        status: 'draft',
        lines: [
          {
            description: data.description || 'Quote',
            quantity: 1,
            unitPrice: data.amount,
            taxRate: 0,
            total: data.amount,
          },
        ],
        subtotal: data.amount,
        tax: 0,
        discount: 0,
        total: data.amount,
        validUntil: data.expiresAt,
        notes: data.description ?? '',
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['quotes'] });
      toast.success(t('quotes.createdToast'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}

export function useUpdateQuoteStatus() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      quoteClient.updateQuoteStatus({ id, body: { status } }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['quotes'] });
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}

export function useConvertQuote() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (id: string) => quoteClient.convertQuote({ id }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['quotes'] });
      void qc.invalidateQueries({ queryKey: ['invoices'] });
      toast.success(t('quotes.convertedToast'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}
