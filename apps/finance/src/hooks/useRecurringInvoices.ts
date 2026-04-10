import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useRecurringInvoices() {
  return useQuery({
    queryKey: ['recurring-invoices'],
    queryFn: () => financeApi.getRecurringInvoices(),
  });
}
