import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useQuotes() {
  return useQuery({
    queryKey: ['quotes'],
    queryFn: () => financeApi.getQuotes(),
  });
}
