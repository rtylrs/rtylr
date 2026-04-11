import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';

export function useBills() {
  return useQuery({
    queryKey: ['bills'],
    queryFn: () => unwrap(financeApi.getBills()),
  });
}
