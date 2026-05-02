import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { financeApi, type TaxConfig } from '../services/financeApi';
import { chartOfAccountSchema } from '../schemas/accounts';
import { journalEntrySchema } from '../schemas/journal';
import type { CreateAccountRequest } from '@voxire/contracts/finance/v1/accounts';
import type { CreateJournalEntryRequest } from '@voxire/contracts/finance/v1/journal';

export function useChartOfAccounts() {
  return useQuery({
    queryKey: ['chart-of-accounts'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getAccounts());
      const raw = res ?? [];
      return z.array(chartOfAccountSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateAccountRequest) => unwrap(financeApi.createAccount(data)),
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
      const res = await unwrap(financeApi.getJournal());
      const raw = res ?? [];
      return z.array(journalEntrySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateJournalEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateJournalEntryRequest) => unwrap(financeApi.createJournalEntry(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['journal'] });
    },
    onError: () => {
      toast.error('Failed to create journal entry');
    },
  });
}

// Re-export TaxConfig so components that need it don't import financeApi directly
export type { TaxConfig };
