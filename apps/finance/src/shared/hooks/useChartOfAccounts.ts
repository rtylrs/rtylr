import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { AccountServiceClient } from '@voxire/voxie/finance/account/v1/client';
import { JournalServiceClient } from '@voxire/voxie/finance/journal/v1/client';
import type { CreateJournalEntryResponse } from '@voxire/voxie/finance/journal/v1/types';
import { chartOfAccountSchema } from '../schemas/accounts';
import { journalEntrySchema } from '../schemas/journal';
import { z } from 'zod';

const accountClient = createOneKitClient(AccountServiceClient, 'finance');
const journalClient = createOneKitClient(JournalServiceClient, 'finance');

export function useChartOfAccounts() {
  return useQuery({
    queryKey: ['chart-of-accounts'],
    queryFn: async () => {
      const res = await accountClient.listAccounts({});
      return z.array(chartOfAccountSchema).parse(res.accounts);
    },
  });
}

export function useCreateAccount() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (data: {
      name: string;
      type: string;
      currency: string;
      balance: number;
      isActive: boolean;
      code?: string;
    }) =>
      accountClient.createAccount({
        name: data.name,
        code: data.code ?? '',
        type: data.type,
        currency: data.currency,
        balance: data.balance,
        isActive: data.isActive,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['chart-of-accounts'] });
      toast.success(t('accounts.created'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
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
    queryKey: ['journal', 'filter', { page, pageSize }],
    queryFn: async (): Promise<JournalResult> => {
      const res = await journalClient.listJournalEntries({
        page,
        limit: pageSize,
        sortBy: '',
        sortDir: '',
        fromDate: '',
        toDate: '',
      });
      const entries = z.array(journalEntrySchema).parse(
        (res.entries ?? []).map((e) => ({
          id: e.id,
          date: e.date,
          description: e.description,
          lines: (e.lines ?? []).map((l) => ({
            account: l.accountName,
            debit: l.debit,
            credit: l.credit,
            description: l.description,
          })),
        }))
      );
      const total = res.pagination?.total ?? entries.length;
      return {
        entries,
        hasMore: page * pageSize < total,
        page,
        pageSize,
      };
    },
  });
}

export function useCreateJournalEntry() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: (data: {
      date: string;
      description: string;
      lines: Array<{
        accountId: number;
        accountName: string;
        debit: number;
        credit: number;
        description: string;
      }>;
    }): Promise<CreateJournalEntryResponse> =>
      journalClient.createJournalEntry({
        date: data.date,
        description: data.description,
        lines: data.lines.map((l) => ({
          accountId: String(l.accountId),
          accountName: l.accountName,
          debit: l.debit,
          credit: l.credit,
          description: l.description,
        })),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['journal'] });
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}

// Re-export TaxConfig so components that need it don't import financeApi directly
export type { TaxConfig } from '../services/financeApi';
