import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { expenseSchema } from '../../shared/schemas/expenses';

export function useExpenseById(id: string | undefined) {
  return useQuery({
    queryKey: ['expense-detail', id],
    queryFn: async () => {
      const res = await unwrap(financeApi.getExpenseDetail(id!));
      return expenseSchema.parse(res);
    },
    enabled: !!id,
  });
}
