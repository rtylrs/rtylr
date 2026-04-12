import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { recurringInvoiceSchema } from '../../shared/schemas/recurringInvoices';

export function useRecurringInvoices() {
  return useQuery({
    queryKey: ['recurring-invoices'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getRecurringInvoices());
      const raw = res ?? [];
      return z.array(recurringInvoiceSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
