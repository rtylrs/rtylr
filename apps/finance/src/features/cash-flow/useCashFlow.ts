import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';

export function useCashFlow() {
  return useQuery({
    queryKey: ['cash-flow'],
    queryFn: () => unwrap(financeApi.getCashFlow()),
  });
}
