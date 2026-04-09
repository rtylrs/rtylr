import { useQuery } from '@tanstack/react-query';
import { buildMockExpense, ExpenseDetail } from '../lib/mockData';

export function useExpenseById(id: string | undefined) {
  return useQuery<ExpenseDetail>({
    queryKey: ['expense-detail', id],
    queryFn: async () => buildMockExpense(id!),
    enabled: !!id,
  });
}
