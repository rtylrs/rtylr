import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';

export function useQuotes() {
  return useQuery({
    queryKey: ['quotes'],
    queryFn: () => unwrap(financeApi.getQuotes()),
  });
}
