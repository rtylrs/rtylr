import { createOneKitClient, unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../lib/api';
import { BankAccountServiceClient } from '@voxire/voxie/finance/bank_account/v1/client';
import { bankAccountSchema } from '../schemas/accounts';
import { unwrapList } from '../lib/unwrapList';

const bankAccountClient = createOneKitClient(BankAccountServiceClient, 'finance');

export function useBankAccounts() {
  return useQuery({
    queryKey: ['bank-accounts'],
    queryFn: async () => {
      const res = await bankAccountClient.listBankAccounts({
        page: 1,
        limit: 100,
        sortBy: '',
        sortDir: '',
      });
      return z.array(bankAccountSchema).parse(
        (res.accounts ?? []).map((a) => ({
          id: a.id,
          name: a.name,
          type: a.accountType,
          balance: a.balance,
          currency: a.currency,
          institution: a.bankName,
          mask: a.lastFour,
        }))
      );
    },
  });
}

// finance.bank_account.v1.BankAccountService has no transaction/feed
// endpoint yet - this genuinely has no backend contract to migrate to.
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
