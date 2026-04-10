import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useBills() {
  return useQuery({
    queryKey: ['bills'],
    queryFn: () => financeApi.getBills(),
  });
}
