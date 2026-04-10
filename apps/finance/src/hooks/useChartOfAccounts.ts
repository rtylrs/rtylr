import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useChartOfAccounts() {
  return useQuery({
    queryKey: ['chart-of-accounts'],
    queryFn: () => financeApi.getAccounts(),
  });
}

export function useJournal() {
  return useQuery({
    queryKey: ['journal'],
    queryFn: () => financeApi.getJournal(),
  });
}
