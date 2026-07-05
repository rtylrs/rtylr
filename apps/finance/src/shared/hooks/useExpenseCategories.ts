import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/service_client';
import type { ExpenseCategory } from '../services/financeApi';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

export function useExpenseCategories() {
  return useQuery<ExpenseCategory[]>({
    queryKey: ['expense-categories'],
    queryFn: async (): Promise<ExpenseCategory[]> => {
      const res = await expensesClient.listExpenseCategories({});
      return res.categories.map((c) => ({ id: Number(c.id), name: c.name, color: c.color }));
    },
  });
}
