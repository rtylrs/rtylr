import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../lib/api';
import { chartOfAccountSchema } from '../schemas/accounts';
import { journalEntrySchema } from '../schemas/journal';
import type { CreateAccountRequest } from '@voxire/contracts/finance/v1/accounts';
import type { CreateJournalEntryRequest } from '@voxire/contracts/finance/v1/journal';
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
    mutationFn: (data: CreateAccountRequest) =>
      unwrap(safe.post('/finance/chart-of-accounts', data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['chart-of-accounts'] });
      toast.success('Account created');
    },
    onError: () => {
      toast.error('Failed to create account');
    },
  });
}

export function useJournal() {
  return useQuery({
    queryKey: ['journal'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/journal'));
      const raw = unwrapList(res, 'entries', 'journal');
      return z.array(journalEntrySchema).parse(Array.isArray(raw) ? raw : []);
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
