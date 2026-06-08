import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import { quoteSchema } from '../../shared/schemas/quotes';
import { unwrapList } from '../../shared/lib/unwrapList';

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
    queryKey: ['quotes', { page, pageSize }],
    queryFn: async (): Promise<QuoteListResult> => {
      const res = await unwrap(
        safe.get<unknown>('/finance/quotes', { params: { page, page_size: pageSize } })
      );
      const raw = res as Record<string, unknown>;
      const quotes = z.array(quoteSchema).parse(unwrapList(res, 'quotes'));
      return {
        quotes,
        hasMore: Boolean(raw.has_more),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.page_size ?? pageSize),
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

interface BackendCreateQuotePayload {
  customerName: string;
  customerEmail: string;
  status: string;
  items: string;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  validUntil: string;
  notes: string;
}

const toBackendCreateQuotePayload = (data: CreateQuotePayload): BackendCreateQuotePayload => ({
  customerName: data.client,
  customerEmail: '',
  status: 'draft',
  items: JSON.stringify([
    {
      description: data.description || 'Quote',
      quantity: 1,
      unit_price: data.amount,
      total: data.amount,
    },
  ]),
  subtotal: data.amount,
  tax: 0,
  discount: 0,
  total: data.amount,
  validUntil: data.expiresAt,
  notes: data.description ?? '',
});

export function useCreateQuote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: CreateQuotePayload): Promise<unknown> =>
      unwrap(safe.post('/finance/quotes', toBackendCreateQuotePayload(data))),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['quotes'] });
      toast.success('Quote created');
    },
    onError: () => {
      toast.error('Failed to create quote');
    },
  });
}

export function useUpdateQuoteStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, status }: { id: string; status: string }): Promise<unknown> =>
      unwrap(safe.put(`/finance/quotes/${id}/status`, { status })),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['quotes'] });
    },
    onError: () => {
      toast.error('Failed to update quote');
    },
  });
}

export function useConvertQuote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string): Promise<unknown> =>
      unwrap(safe.post(`/finance/quotes/${id}/convert`, {})),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['quotes'] });
      void qc.invalidateQueries({ queryKey: ['invoices'] });
      toast.success('Quote converted to invoice');
    },
    onError: () => {
      toast.error('Failed to convert quote');
    },
  });
}
