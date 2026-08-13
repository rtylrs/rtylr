import { createOneKitClient, getOneKitErrorMessage, toFiniteNumber } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import { ReportServiceClient } from '@voxire/voxie/finance/report/v1/client';
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
    mutationFn: async (data: { name: string; color: string; icon?: string }) => {
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
