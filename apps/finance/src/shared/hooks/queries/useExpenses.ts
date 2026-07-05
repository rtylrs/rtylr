import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { Expense } from '@voxire/voxie/finance/expenses/v1/types';
import type { ListExpensesResponse } from '@voxire/voxie/finance/expenses/v1/list_expenses';
import type {
  CreateExpenseRequest,
  CreateExpenseResponse,
} from '@voxire/voxie/finance/expenses/v1/create_expense';
import type {
  UpdateExpenseRequest,
  UpdateExpenseResponse,
} from '@voxire/voxie/finance/expenses/v1/update_expense';
import type { DeleteExpenseResponse } from '@voxire/voxie/finance/expenses/v1/delete_expense';
import type { ListExpenseCategoriesResponse } from '@voxire/voxie/finance/expenses/v1/list_expense_categories';
import { safe } from '../../lib/api';

interface ListExpensesParams {
  limit: number;
  offset: number;
  category?: string;
  start?: string;
  end?: string;
}

export const useListExpenses = (params: ListExpensesParams) => {
  return useQuery<ListExpensesResponse>({
    queryKey: ['expenses', 'filter', params],
    queryFn: async (): Promise<ListExpensesResponse> => {
      const res = await unwrap(safe.get<ListExpensesResponse>('/finance/expenses', { params }));
      return res as ListExpensesResponse;
    },
  });
};

export const useGetExpense = (id: number | undefined) => {
  return useQuery<Expense>({
    queryKey: ['expense', id],
    queryFn: async (): Promise<Expense> => {
      const res = await unwrap(safe.get<{ expense: Expense }>(`/finance/expenses/${id}`));
      return (res as { expense: Expense }).expense;
    },
    enabled: !!id,
  });
};

export const useExpenseCategories = () => {
  return useQuery<ListExpenseCategoriesResponse>({
    queryKey: ['expense-categories'],
    queryFn: async (): Promise<ListExpenseCategoriesResponse> => {
      const res = await unwrap(
        safe.get<ListExpenseCategoriesResponse>('/finance/expenses/categories')
      );
      return res as ListExpenseCategoriesResponse;
    },
  });
};

export const useCreateExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateExpenseRequest): Promise<CreateExpenseResponse> => {
      const res = await unwrap(safe.post<CreateExpenseResponse>('/finance/expenses', data));
      return res as CreateExpenseResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      toast.success(t('expenses.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

interface UpdateExpenseVariables extends UpdateExpenseRequest {
  id: number;
}

export const useUpdateExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpdateExpenseVariables): Promise<UpdateExpenseResponse> => {
      const { id, ...body } = data;
      const res = await unwrap(safe.put<UpdateExpenseResponse>(`/finance/expenses/${id}`, body));
      return res as UpdateExpenseResponse;
    },
    onSuccess: (_data: UpdateExpenseResponse, variables: UpdateExpenseVariables) => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      queryClient.invalidateQueries({ queryKey: ['expense', variables.id] });
      toast.success(t('expenses.updated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: number): Promise<DeleteExpenseResponse> => {
      const res = await unwrap(safe.delete<DeleteExpenseResponse>(`/finance/expenses/${id}`));
      return res as DeleteExpenseResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      toast.success(t('expenses.deleted'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApproveExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: string): Promise<void> => {
      await unwrap(safe.post<void>(`/finance/expenses/${id}/approve`, {}));
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      queryClient.invalidateQueries({ queryKey: ['expense', id] });
      toast.success(t('expenses.approved'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useRejectExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: string): Promise<void> => {
      await unwrap(safe.post<void>(`/finance/expenses/${id}/reject`, {}));
    },
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['expenses'] });
      queryClient.invalidateQueries({ queryKey: ['expense', id] });
      toast.success(t('expenses.rejected'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
