import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useCashFlow() {
  return useQuery({
    queryKey: ['cash-flow'],
    queryFn: () => financeApi.getCashFlow(),
  });
}
