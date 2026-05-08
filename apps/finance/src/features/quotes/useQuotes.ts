import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import { quoteSchema } from '../../shared/schemas/quotes';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useQuotes() {
  return useQuery({
    queryKey: ['quotes'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/quotes'));
      return z.array(quoteSchema).parse(unwrapList(res, 'quotes'));
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
  return useMutation({
    mutationFn: async (data: CreateQuotePayload): Promise<unknown> =>
      unwrap(safe.post('/finance/quotes', data)),
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
      unwrap(safe.put(`/finance/quotes/${id}`, { status })),
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
