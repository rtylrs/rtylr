import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ListJournalEntriesResponse } from '@voxire/voxie/finance/journal/v1/list_journal_entries';
import type {
  CreateJournalEntryRequest,
  CreateJournalEntryResponse,
} from '@voxire/voxie/finance/journal/v1/create_journal_entry';
import { safe } from '../../lib/api';

interface ListJournalParams {
  page?: number;
  pageSize?: number;
  fromDate?: string;
  toDate?: string;
}

export interface JournalListResult {
  entries: ListJournalEntriesResponse['entries'];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export const useListJournal = (params: ListJournalParams = {}) => {
  const { page = 1, pageSize = 50, fromDate, toDate } = params;
  return useQuery<JournalListResult>({
    queryKey: ['journal', 'filter', { page, pageSize, fromDate, toDate }],
    queryFn: async (): Promise<JournalListResult> => {
      const res = await unwrap(
        safe.get<unknown>('/finance/journal', {
          params: { page, page_size: pageSize, from_date: fromDate, to_date: toDate },
        })
      );
      const raw = res as Record<string, unknown>;
      const rawRes = res as ListJournalEntriesResponse;
      return {
        entries: rawRes.entries ?? [],
        hasMore: Boolean(raw.has_more),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.page_size ?? pageSize),
      };
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
