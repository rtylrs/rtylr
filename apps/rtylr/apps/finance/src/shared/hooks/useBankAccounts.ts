import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { financeApi } from '../services/financeApi';
import { BankAccountServiceClient } from '@voxire/voxie/finance/bank_account/v1/client';
import { bankAccountSchema, bankTransactionSchema } from '../schemas/accounts';
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

export function useCreateBankAccount() {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: {
      name: string;
      bankName: string;
      accountType: string;
      lastFour: string;
      balance: number;
      currency: string;
    }) =>
      bankAccountClient.createBankAccount({
        name: data.name,
        bankName: data.bankName,
        accountType: data.accountType,
        lastFour: data.lastFour,
        balance: data.balance,
        currency: data.currency,
        isPrimary: false,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['bank-accounts'] });
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}

// finance.bank_account.v1.BankAccountService has no transaction/feed
// endpoint yet - this genuinely has no backend contract to migrate to.
export function useBankTransactions(accountId?: string) {
  return useQuery({
    queryKey: ['bank-tx', accountId ?? 'all'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getBankFeeds(accountId));
      return z.array(bankTransactionSchema).parse(unwrapList(res, 'transactions'));
    },
  });
}
