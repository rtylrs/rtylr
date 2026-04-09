import { useQuery } from '@tanstack/react-query';
import { MOCK_BANK_ACCOUNTS, MOCK_BANK_TX, BankAccount, BankTransaction } from '../lib/mockData';

export function useBankAccounts() {
  return useQuery<BankAccount[]>({
    queryKey: ['bank-accounts'],
    queryFn: async () => MOCK_BANK_ACCOUNTS,
  });
}

export function useBankTransactions(accountId?: string) {
  return useQuery<BankTransaction[]>({
    queryKey: ['bank-tx', accountId ?? 'all'],
    queryFn: async () => (accountId ? MOCK_BANK_TX.filter((t) => t.accountId === accountId) : MOCK_BANK_TX),
  });
}
