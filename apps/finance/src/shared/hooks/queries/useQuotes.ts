import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ListQuotesResponse } from '@voxire/voxie/finance/quote/v1/list_quotes';
import type { CreateQuoteResponse } from '@voxire/voxie/finance/quote/v1/create_quote';
import { safe } from '../../lib/api';

interface ListQuotesParams {
  limit: number;
  offset: number;
  status?: string;
  customerId?: number;
}

export interface CreateQuotePayload {
  customerName: string;
  customerEmail?: string;
  status?: string;
  items?: string;
  subtotal?: number;
  tax?: number;
  discount?: number;
  total?: number;
  validUntil?: string;
  notes?: string;
}

export const useListQuotes = (params: ListQuotesParams) => {
  return useQuery<ListQuotesResponse>({
    queryKey: ['quotes', 'filter', params],
    queryFn: async (): Promise<ListQuotesResponse> => {
      const res = await unwrap(safe.get<ListQuotesResponse>('/finance/quotes', { params }));
      return res as ListQuotesResponse;
    },
  });
};

export const useCreateQuote = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateQuotePayload): Promise<CreateQuoteResponse> => {
      const res = await unwrap(safe.post<CreateQuoteResponse>('/finance/quotes', data));
      return res as CreateQuoteResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['quotes'] });
      toast.success(t('quotes.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
