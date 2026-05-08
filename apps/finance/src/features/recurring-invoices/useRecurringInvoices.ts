import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import { recurringInvoiceSchema } from '../../shared/schemas/recurringInvoices';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useRecurringInvoices() {
  return useQuery({
    queryKey: ['recurring-invoices'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/recurring-invoices'));
      return z.array(recurringInvoiceSchema).parse(unwrapList(res, 'recurring_invoices', 'items'));
    },
  });
}

interface CreateRecurringPayload {
  client: string;
  amount: number;
  currency: string;
  cadence: string;
  startDate: string;
  description?: string;
}

export function useCreateRecurringInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateRecurringPayload): Promise<unknown> =>
      unwrap(safe.post('/finance/recurring-invoices', data as unknown as object)),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['recurring-invoices'] });
    },
  });
}

export function useSetRecurringStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'active' | 'paused' }): Promise<void> =>
      unwrap(safe.put<void>(`/finance/recurring-invoices/${id}`, { status } as unknown as object)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recurring-invoices'] }),
  });
}
