import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../lib/api';
import { bankAccountSchema } from '../schemas/accounts';
import { unwrapList } from '../lib/unwrapList';

export function useBankAccounts() {
  return useQuery({
    queryKey: ['bank-accounts'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/bank-accounts'));
      return z.array(bankAccountSchema).parse(unwrapList(res, 'accounts', 'bank_accounts'));
    },
  });
}

export function useBankTransactions(accountId?: string) {
  return useQuery({
    queryKey: ['bank-tx', accountId ?? 'all'],
    queryFn: async () => {
      const params = accountId ? { account_id: accountId } : undefined;
      const res = await unwrap(safe.get<unknown>('/finance/bank-feeds', { params }));
      return z.array(bankAccountSchema).parse(unwrapList(res, 'transactions'));
    },
  });
}
