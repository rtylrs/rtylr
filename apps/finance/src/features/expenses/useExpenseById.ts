import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';

export function useExpenseById(id: string | undefined) {
  return useQuery({
    queryKey: ['expense-detail', id],
    queryFn: () => unwrap(financeApi.getExpenseDetail(id!)),
    enabled: !!id,
  });
}
