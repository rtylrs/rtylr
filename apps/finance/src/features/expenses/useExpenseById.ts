import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import type { Expense } from '../../shared/services/financeApi';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

export function useExpenseById(id: string | undefined) {
  return useQuery<Expense | undefined>({
    queryKey: ['expense', id],
    queryFn: async (): Promise<Expense | undefined> => {
      const res = await expensesClient.getExpense({ id: id ?? '' });
      if (!res.expense) return undefined;
      return {
        id: res.expense.id ?? '',
        categoryId: res.expense.categoryId != null ? toFiniteNumber(res.expense.categoryId) : null,
        description: res.expense.description ?? '',
        amount: toFiniteNumber(res.expense.amount),
        currency: res.expense.currency,
        vendor: res.expense.vendor,
        receiptUrl: res.expense.receiptUrl,
        expenseDate: res.expense.expenseDate ?? '',
        status: res.expense.status ?? '',
        createdAt: res.expense.createdAt,
        updatedAt: res.expense.updatedAt,
      };
    },
    enabled: !!id,
  });
}
