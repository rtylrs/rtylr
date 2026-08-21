import { createOneKitClient, getOneKitErrorMessage, toFiniteNumber, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import { ReportServiceClient } from '@voxire/voxie/finance/report/v1/client';
import { financeApi } from '../services/financeApi';
import type { ExpenseCategory } from '../services/financeApi';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');
const reportClient = createOneKitClient(ReportServiceClient, 'finance');

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

export function useCreateExpenseCategory() {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: { name: string; color: string; icon?: string; budget?: number }) => {
      const response = await reportClient.createExpenseCategoryDetail({
        name: data.name,
        color: data.color,
        icon: data.icon ?? 'tag',
      });
      return response.category;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['expense-categories'] });
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}

export function useUpdateExpenseCategory() {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async ({
      id,
      data,
    }: {
      id: number;
      data: { name: string; color: string; budget?: number; icon?: string };
    }) => {
      return unwrap(
        financeApi.updateExpenseCategory(id, {
          name: data.name,
          color: data.color,
          budget: data.budget,
          icon: data.icon ?? 'tag',
        })
      );
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['expense-categories'] });
      toast.success(t('categories.updated', 'Category updated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}

export function useDeleteExpenseCategory() {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: number) => {
      return unwrap(financeApi.deleteExpenseCategory(id));
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['expense-categories'] });
      toast.success(t('categories.deleted', 'Category deleted'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
}
