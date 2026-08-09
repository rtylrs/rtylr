import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import type { ExpenseCategory } from '../services/financeApi';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

export function useExpenseCategories() {
  return useQuery<ExpenseCategory[]>({
    queryKey: ['expense-categories'],
    queryFn: async (): Promise<ExpenseCategory[]> => {
      const res = await expensesClient.listExpenseCategories({});
      return (res.categories ?? []).map((c) => ({
        id: toFiniteNumber(c.id),
        name: c.name ?? '',
        color: c.color ?? '',
      }));
    },
  });
}
