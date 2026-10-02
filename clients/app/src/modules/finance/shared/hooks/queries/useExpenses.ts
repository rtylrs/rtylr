import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

export const useApproveExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    mutationFn: (id: string): Promise<void> =>
      expensesClient
        .updateExpenseStatus({ id, body: { status: 'approved' } })
        .then(() => undefined),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['finance', 'expenses'] });
      queryClient.invalidateQueries({ queryKey: ['finance', 'expense', id] });
      toast.success(t('financeExpenses.approved'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
};

export const useRejectExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    mutationFn: (id: string): Promise<void> =>
      expensesClient
        .updateExpenseStatus({ id, body: { status: 'rejected' } })
        .then(() => undefined),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['finance', 'expenses'] });
      queryClient.invalidateQueries({ queryKey: ['finance', 'expense', id] });
      toast.success(t('financeExpenses.rejected'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
};
