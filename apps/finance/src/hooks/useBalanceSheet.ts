import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useBalanceSheet() {
  return useQuery({
    queryKey: ['balance-sheet'],
    queryFn: () => financeApi.getBalanceSheet(),
  });
}
