import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type {
  ListJournalEntriesResponse,
  CreateJournalEntryRequest,
  CreateJournalEntryResponse,
} from '@voxire/contracts/finance/v1/journal';
import { safe } from '../../lib/api';

interface ListJournalParams {
  limit: number;
  offset: number;
  fromDate?: string;
  toDate?: string;
}

export const useListJournal = (params: ListJournalParams) => {
  return useQuery<ListJournalEntriesResponse>({
    queryKey: ['journal', 'filter', params],
    queryFn: async (): Promise<ListJournalEntriesResponse> => {
      const res = await unwrap(
        safe.get<ListJournalEntriesResponse>('/finance/journal', { params })
      );
      return res as ListJournalEntriesResponse;
    },
  });
};

export const useCreateJournalEntry = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateJournalEntryRequest): Promise<CreateJournalEntryResponse> => {
      const res = await unwrap(safe.post<CreateJournalEntryResponse>('/finance/journal', data));
      return res as CreateJournalEntryResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['journal'] });
      toast.success(t('journal.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
