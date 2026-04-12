import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financeApi, Expense, ExpenseCategory } from '../../shared/services/financeApi';
import { expenseSchema, expenseCategorySchema } from '../../shared/schemas/expenses';

export function useExpenses() {
  return useQuery({
    queryKey: ['expenses'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getExpenses());
      const raw = res ?? [];
      return z.array(expenseSchema).parse(Array.isArray(raw) ? raw : []) as Expense[];
    },
  });
}

export function useExpenseCategories() {
  return useQuery({
    queryKey: ['expense-categories'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getExpenseCategories());
      const raw = res ?? [];
      return z
        .array(expenseCategorySchema)
        .parse(Array.isArray(raw) ? raw : []) as ExpenseCategory[];
    },
  });
}

export function useCreateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Expense>) => unwrap(financeApi.createExpense(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
  });
}

export function useUpdateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Expense> }) =>
      unwrap(financeApi.updateExpense(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
  });
}

export function useDeleteExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(financeApi.deleteExpense(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
  });
}

export function useCreateExpenseCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ExpenseCategory>) => unwrap(financeApi.createExpenseCategory(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expense-categories'] }),
  });
}
