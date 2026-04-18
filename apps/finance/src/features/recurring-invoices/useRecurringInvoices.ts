import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { recurringInvoiceSchema } from '../../shared/schemas/recurringInvoices';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useRecurringInvoices() {
  return useQuery({
    queryKey: ['recurring-invoices'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getRecurringInvoices());
      return z.array(recurringInvoiceSchema).parse(unwrapList(res, 'recurring_invoices', 'items'));
    },
  });
}
