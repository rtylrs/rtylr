import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useFinanceDashboard() {
  return useQuery({
    queryKey: ['finance-dashboard'],
    queryFn: () => financeApi.getDashboard(),
  });
}
