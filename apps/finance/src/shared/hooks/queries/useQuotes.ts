import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type {
  ListQuotesResponse,
  CreateQuoteRequest,
  CreateQuoteResponse,
} from '@voxire/contracts/finance/v1/quotes';
import { safe } from '../../lib/api';

interface ListQuotesParams {
  limit: number;
  offset: number;
  status?: string;
  customerId?: number;
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
    mutationFn: async (data: CreateQuoteRequest): Promise<CreateQuoteResponse> => {
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
