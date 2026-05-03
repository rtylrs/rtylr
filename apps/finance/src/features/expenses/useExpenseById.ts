import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import { expenseSchema } from '../../shared/schemas/expenses';

export function useExpenseById(id: string | undefined) {
  return useQuery({
    queryKey: ['expense', id],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>(`/finance/expenses/${id}`));
      return expenseSchema.parse(res);
    },
    enabled: !!id,
  });
}
