import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import type { Expense, ExpenseCategory } from '../../shared/services/financeApi';
import { expenseSchema, expenseCategorySchema } from '../../shared/schemas/expenses';
import { unwrapList } from '../../shared/lib/unwrapList';

interface UseExpensesParams {
  page?: number;
  pageSize?: number;
}

export interface ExpenseListResult {
  expenses: Expense[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export function useExpenses(params: UseExpensesParams = {}) {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<ExpenseListResult>({
    queryKey: ['expenses', { page, pageSize }],
    queryFn: async (): Promise<ExpenseListResult> => {
      const res = await unwrap(
        safe.get<unknown>('/finance/expenses', { params: { page, page_size: pageSize } })
      );
      const raw = res as Record<string, unknown>;
      const expenses = z.array(expenseSchema).parse(unwrapList(res, 'expenses')) as Expense[];
      return {
        expenses,
        hasMore: Boolean(raw.has_more),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.page_size ?? pageSize),
      };
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
    onError: () => toast.error('Failed to create expense'),
  });
}

export function useUpdateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Expense> }): Promise<Expense> =>
      unwrap(safe.put<Expense>(`/finance/expenses/${id}`, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: () => toast.error('Failed to update expense'),
  });
}

export function useDeleteExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string): Promise<void> => unwrap(safe.delete<void>(`/finance/expenses/${id}`)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: () => toast.error('Failed to delete expense'),
  });
}

export function useCreateExpenseCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ExpenseCategory>): Promise<ExpenseCategory> =>
      unwrap(safe.post<ExpenseCategory>('/finance/expense-categories', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expense-categories'] }),
    onError: () => toast.error('Failed to create expense category'),
  });
}
