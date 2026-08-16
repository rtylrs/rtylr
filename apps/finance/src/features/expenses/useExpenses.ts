import { createOneKitClient, getOneKitErrorMessage, toFiniteNumber } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import type {
  Expense as ApiExpense,
  ExpenseCategory as ApiExpenseCategory,
} from '@voxire/voxie/finance/expenses/v1/types';
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
    id: e.id ?? '',
    categoryId: e.categoryId != null ? toFiniteNumber(e.categoryId) : null,
    description: e.description ?? '',
    amount: toFiniteNumber(e.amount),
    currency: e.currency,
    vendor: e.vendor,
    receiptUrl: e.receiptUrl,
    expenseDate: e.expenseDate ?? '',
    status: e.status ?? '',
    createdAt: e.createdAt,
    updatedAt: e.updatedAt,
  };
}

function fromApiCategory(c: ApiExpenseCategory): ExpenseCategory {
  return { id: toFiniteNumber(c.id), name: c.name ?? '', color: c.color ?? '' };
}

export function useExpenses(params: UseExpensesParams = {}) {
  const { page = 1, pageSize = 25 } = params;
  return useQuery<ExpenseListResult>({
    queryKey: ['expenses', 'filter', { page, pageSize }],
    queryFn: async (): Promise<ExpenseListResult> => {
      const res = await expensesClient.listExpenses({});
      const all = (res.expenses ?? []).map(fromApiExpense);
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
      return (res.categories ?? []).map(fromApiCategory);
    },
  });
}

export function useCreateExpense() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (data: Partial<Expense>): Promise<{ expenseId: string; message: string }> => {
      const res = await expensesClient.createExpense({
        categoryId: data.categoryId != null ? String(data.categoryId) : undefined,
        amount: data.amount ?? 0,
        currency: data.currency ?? 'USD',
        description: data.description ?? '',
        vendor: data.vendor ?? '',
        receiptUrl: data.receiptUrl ?? '',
        expenseDate: data.expenseDate ?? '',
      });
      return { expenseId: res.expenseId ?? '', message: res.message ?? '' };
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

export function useUpdateExpense() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async ({
      id,
      data,
    }: {
      id: string;
      data: Partial<Expense>;
    }): Promise<{ message: string }> => {
      const res = await expensesClient.updateExpense({
        id,
        body: {
          categoryId: data.categoryId != null ? String(data.categoryId) : undefined,
          amount: data.amount ?? 0,
          currency: data.currency ?? 'USD',
          description: data.description ?? '',
          vendor: data.vendor ?? '',
          receiptUrl: data.receiptUrl ?? '',
          expenseDate: data.expenseDate ?? '',
        },
      });
      return { message: res.message ?? '' };
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}

export function useDeleteExpense() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (id: string): Promise<{ message: string }> => {
      const res = await expensesClient.deleteExpense({ id });
      return { message: res.message ?? '' };
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['expenses'] }),
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('errors.generic'))),
  });
}
