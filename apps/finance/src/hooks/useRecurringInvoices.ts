import { useQuery } from '@tanstack/react-query';
import { MOCK_RECURRING, RecurringInvoice } from '../lib/mockData';

export function useRecurringInvoices() {
  return useQuery<RecurringInvoice[]>({
    queryKey: ['recurring-invoices'],
    queryFn: async () => MOCK_RECURRING,
  });
}
