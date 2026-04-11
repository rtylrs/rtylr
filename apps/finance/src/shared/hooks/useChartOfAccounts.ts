import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useChartOfAccounts() {
  return useQuery({
    queryKey: ['chart-of-accounts'],
    queryFn: () => unwrap(financeApi.getAccounts()),
  });
}

export function useJournal() {
  return useQuery({
    queryKey: ['journal'],
    queryFn: () => unwrap(financeApi.getJournal()),
  });
}
