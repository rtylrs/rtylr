import { createOneKitClient, getOneKitErrorMessage, toFiniteNumber } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  ExpensesServiceClient,
  type Expense as ApiExpense,
  type ExpenseCategory as ApiExpenseCategory,
} from '@voxire/voxie/finance/expenses/v1/service_client';
import type { Expense, ExpenseCategory } from '../../shared/services/financeApi';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

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

function fromApiExpense(e: ApiExpense): Expense {
  return {
    id: e.id,
    categoryId: e.categoryId != null ? toFiniteNumber(e.categoryId) : null,
    description: e.description,
    amount: toFiniteNumber(e.amount),
    currency: e.currency,
    vendor: e.vendor,
    receiptUrl: e.receiptUrl,
    expenseDate: e.expenseDate,
    status: e.status,
    createdAt: e.createdAt,
    updatedAt: e.updatedAt,
  };
}

function fromApiCategory(c: ApiExpenseCategory): ExpenseCategory {
  return { id: toFiniteNumber(c.id), name: c.name, color: c.color };
}

// finance.expenses.v1.ExpensesService.ListExpenses has no pagination
// (returns the full set) - paginate client-side to keep this hook's
// page/pageSize contract intact for callers.
export function useExpenses(params: UseExpensesParams = {}) {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<ExpenseListResult>({
    queryKey: ['expenses', { page, pageSize }],
    queryFn: async (): Promise<ExpenseListResult> => {
      const res = await expensesClient.listExpenses({});
      const all = res.expenses.map(fromApiExpense);
      const start = (page - 1) * pageSize;
      const pageItems = all.slice(start, start + pageSize);
      return {
        expenses: pageItems,
        hasMore: start + pageSize < all.length,
        page,
        pageSize,
      };
    },
  });
}

export function useExpenseCategories() {
  return useQuery<ExpenseCategory[]>({
    queryKey: ['expense-categories'],
    queryFn: async (): Promise<ExpenseCategory[]> => {
      const res = await expensesClient.listExpenseCategories({});
      return res.categories.map(fromApiCategory);
    },
  });
}

export function useCreateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Expense>): Promise<{ expenseId: string; message: string }> =>
      expensesClient.createExpense({
        categoryId: data.categoryId != null ? String(data.categoryId) : null,
        amount: data.amount ?? 0,
        currency: data.currency ?? 'USD',
        description: data.description ?? '',
        vendor: data.vendor ?? '',
        receiptUrl: data.receiptUrl ?? '',
        expenseDate: data.expenseDate ?? '',
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to create expense')),
  });
}

export function useUpdateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Partial<Expense>;
    }): Promise<{ message: string }> =>
      expensesClient.updateExpense({
        id,
        body: {
          categoryId: data.categoryId != null ? String(data.categoryId) : null,
          amount: data.amount ?? 0,
          currency: data.currency ?? 'USD',
          description: data.description ?? '',
          vendor: data.vendor ?? '',
          receiptUrl: data.receiptUrl ?? '',
          expenseDate: data.expenseDate ?? '',
        },
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to update expense')),
  });
}

export function useDeleteExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string): Promise<{ message: string }> => expensesClient.deleteExpense({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to delete expense')),
  });
}
