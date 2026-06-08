import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../lib/api';
import { chartOfAccountSchema } from '../schemas/accounts';
import { journalEntrySchema } from '../schemas/journal';
import type { CreateAccountRequest } from '@voxire/contracts/finance/account/v1/create_account';
import type { CreateJournalEntryRequest } from '@voxire/contracts/finance/journal/v1/create_journal_entry';
import { unwrapList } from '../lib/unwrapList';

export function useChartOfAccounts() {
  return useQuery({
    queryKey: ['chart-of-accounts'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/chart-of-accounts'));
      const raw = unwrapList(res, 'accounts', 'chart_of_accounts');
      return z.array(chartOfAccountSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateAccountRequest) => unwrap(safe.post('/finance/accounts', data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['chart-of-accounts'] });
      toast.success('Account created');
    },
    onError: () => {
      toast.error('Failed to create account');
    },
  });
}

interface UseJournalParams {
  page?: number;
  pageSize?: number;
}

export interface JournalResult {
  entries: z.infer<typeof journalEntrySchema>[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export function useJournal(params: UseJournalParams = {}) {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<JournalResult>({
    queryKey: ['journal', { page, pageSize }],
    queryFn: async (): Promise<JournalResult> => {
      const res = await unwrap(
        safe.get<unknown>('/finance/journal', { params: { page, page_size: pageSize } })
      );
      const raw = res as Record<string, unknown>;
      const list = unwrapList(res, 'entries', 'journal');
      const entries = z.array(journalEntrySchema).parse(Array.isArray(list) ? list : []);
      return {
        entries,
        hasMore: Boolean(raw.has_more),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.page_size ?? pageSize),
      };
    },
  });
}

export function useCreateJournalEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateJournalEntryRequest) => unwrap(safe.post('/finance/journal', data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['journal'] });
    },
    onError: () => {
      toast.error('Failed to create journal entry');
    },
  });
}

// Re-export TaxConfig so components that need it don't import financeApi directly
export type { TaxConfig } from '../services/financeApi';
