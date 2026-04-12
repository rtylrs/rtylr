import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';
import { bankAccountSchema } from '../schemas/accounts';

export function useBankAccounts() {
  return useQuery({
    queryKey: ['bank-accounts'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getBankAccounts());
      const raw = res ?? [];
      return z.array(bankAccountSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useBankTransactions(accountId?: string) {
  return useQuery({
    queryKey: ['bank-tx', accountId ?? 'all'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getBankAccounts());
      const raw = res ?? [];
      return z.array(bankAccountSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
