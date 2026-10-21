import { createOneKitClient, getOneKitErrorMessage, toFiniteNumber } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import type { ExpenseCategory as ApiExpenseCategory } from '@voxire/voxie/finance/expenses/v1/types';
import type { ExpenseCategory } from '../services/financeApi';

// Every category read and write goes through finance.expenses.v1. Creating
// through finance.report.v1 while listing through expenses left new
// categories invisible behind the expenses module's list cache.
const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

const categoriesKey = ['finance', 'expense-categories'] as const;

const fromApiCategory = (c: ApiExpenseCategory | undefined): ExpenseCategory => ({
  id: toFiniteNumber(c?.id),
  name: c?.name ?? '',
  color: c?.color ?? '',
});

export function useExpenseCategories() {
  return useQuery<ExpenseCategory[]>({
    queryKey: categoriesKey,
    queryFn: async (): Promise<ExpenseCategory[]> => {
      const res = await expensesClient.listExpenseCategories({});
      return (res.categories ?? []).map(fromApiCategory);
    },
  });
}

export function useCreateExpenseCategory() {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    // The category contract carries no budget column, so none is sent.
    mutationFn: async (data: { name: string; color: string }) => {
      const response = await expensesClient.createExpenseCategory({
        name: data.name,
        color: data.color,
      });
      return { category: fromApiCategory(response.category), created: response.created ?? true };
    },
    onSuccess: ({ category, created }) => {
      void queryClient.invalidateQueries({ queryKey: categoriesKey });
      if (created) {
        toast.success(t('financeCategories.created', 'Category created'));
      } else {
        toast.info(
          t('financeCategories.alreadyExists', '“{{name}}” already exists', { name: category.name })
        );
      }
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
}

export function useUpdateExpenseCategory() {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    mutationFn: async ({ id, data }: { id: number; data: { name: string; color: string } }) => {
      const response = await expensesClient.updateExpenseCategory({
        id: String(id),
        body: { name: data.name, color: data.color },
      });
      return fromApiCategory(response.category);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: categoriesKey });
      toast.success(t('financeCategories.updated', 'Category updated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
}

export function useDeleteExpenseCategory() {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    mutationFn: async (id: number) => {
      await expensesClient.deleteExpenseCategory({ id: String(id) });
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: categoriesKey });
      toast.success(t('financeCategories.deleted', 'Category deleted'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
}
