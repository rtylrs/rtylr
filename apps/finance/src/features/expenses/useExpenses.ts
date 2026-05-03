import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import type { Expense, ExpenseCategory } from '../../shared/services/financeApi';
import { expenseSchema, expenseCategorySchema } from '../../shared/schemas/expenses';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useExpenses() {
  return useQuery<Expense[]>({
    queryKey: ['expenses'],
    queryFn: async (): Promise<Expense[]> => {
      const res = await unwrap(safe.get<unknown>('/finance/expenses'));
      return z.array(expenseSchema).parse(unwrapList(res, 'expenses')) as Expense[];
    },
  });
}

export function useExpenseCategories() {
  return useQuery<ExpenseCategory[]>({
    queryKey: ['expense-categories'],
    queryFn: async (): Promise<ExpenseCategory[]> => {
      const res = await unwrap(safe.get<unknown>('/finance/expenses/categories'));
      return z
        .array(expenseCategorySchema)
        .parse(unwrapList(res, 'categories')) as ExpenseCategory[];
    },
  });
}

export function useCreateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Expense>): Promise<Expense> =>
      unwrap(safe.post<Expense>('/finance/expenses', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
  });
}

export function useUpdateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Expense> }): Promise<Expense> =>
      unwrap(safe.put<Expense>(`/finance/expenses/${id}`, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
  });
}

export function useDeleteExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string): Promise<void> => unwrap(safe.delete<void>(`/finance/expenses/${id}`)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
  });
}

export function useCreateExpenseCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ExpenseCategory>): Promise<ExpenseCategory> =>
      unwrap(safe.post<ExpenseCategory>('/finance/expenses/categories', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expense-categories'] }),
  });
}
