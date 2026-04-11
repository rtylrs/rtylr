import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { financeApi, Expense, ExpenseCategory } from '../../shared/services/financeApi';

export function useExpenses() {
  return useQuery({
    queryKey: ['expenses'],
    queryFn: () => unwrap(financeApi.getExpenses()),
  });
}

export function useExpenseCategories() {
  return useQuery({
    queryKey: ['expense-categories'],
    queryFn: () => unwrap(financeApi.getExpenseCategories()),
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
    mutationFn: ({ id, data }: { id: string; data: Partial<Expense> }) => unwrap(financeApi.updateExpense(id, data)),
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
