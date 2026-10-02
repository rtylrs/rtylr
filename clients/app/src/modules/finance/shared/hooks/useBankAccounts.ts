import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { z } from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { financeApi } from '../services/financeApi';
import { BankAccountServiceClient } from '@voxire/voxie/finance/bank_account/v1/client';
import { bankAccountSchema, bankTransactionSchema } from '../schemas/accounts';

const bankAccountClient = createOneKitClient(BankAccountServiceClient, 'finance');

export function useBankAccounts() {
  return useQuery({
    queryKey: ['finance', 'bank-accounts'],
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
  const { t } = useTranslation('financeCommon');

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
      void queryClient.invalidateQueries({ queryKey: ['finance', 'bank-accounts'] });
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
}

function bankTransactionRows(payload: unknown): unknown[] {
  if (Array.isArray(payload)) return payload;
  if (payload && typeof payload === 'object') {
    const transactions = (payload as { transactions?: unknown }).transactions;
    if (Array.isArray(transactions)) return transactions;
  }
  throw new Error('Invalid bank feed response');
}

// The literal /finance/bank-feeds endpoint is not part of the generated
// BankAccountService contract, but it does have a live response contract. Map
// and validate that response here so failures reach React Query instead of
// being converted into a successful empty list.
export function useBankTransactions(accountId?: string) {
  return useQuery({
    queryKey: ['finance', 'bank-tx', accountId ?? 'all'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getBankFeeds(accountId));
      return z.array(bankTransactionSchema).parse(bankTransactionRows(res));
    },
  });
}
