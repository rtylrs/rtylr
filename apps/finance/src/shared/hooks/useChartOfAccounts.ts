import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';
import { chartOfAccountSchema } from '../schemas/accounts';
import { journalEntrySchema } from '../schemas/journal';

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
