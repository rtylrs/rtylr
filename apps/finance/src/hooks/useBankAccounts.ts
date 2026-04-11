import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useBankAccounts() {
  return useQuery({
    queryKey: ['bank-accounts'],
    queryFn: () => unwrap(financeApi.getBankAccounts()),
  });
}

export function useBankTransactions(accountId?: string) {
  return useQuery({
    queryKey: ['bank-tx', accountId ?? 'all'],
    queryFn: () => unwrap(financeApi.getBankAccounts()),
  });
}
