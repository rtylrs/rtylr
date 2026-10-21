import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

/**
 * Refreshes everything an expense change can move: the lists, the P&L (which
 * counts approved expenses and reports pending ones) and the dashboard.
 */
export function invalidateExpenseQueries(queryClient: QueryClient): void {
  void queryClient.invalidateQueries({ queryKey: ['finance', 'expenses'] });
  void queryClient.invalidateQueries({ queryKey: ['finance', 'pl'] });
  void queryClient.invalidateQueries({ queryKey: ['finance', 'finance-dashboard'] });
}

export const useApproveExpense = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    mutationFn: (id: string): Promise<void> =>
      expensesClient
        .updateExpenseStatus({ id, body: { status: 'approved' } })
        .then(() => undefined),
    onSuccess: (_data, id) => {
      invalidateExpenseQueries(queryClient);
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
      invalidateExpenseQueries(queryClient);
      queryClient.invalidateQueries({ queryKey: ['finance', 'expense', id] });
      toast.success(t('financeExpenses.rejected'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
};
