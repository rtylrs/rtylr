import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import { recurringInvoiceSchema } from '../../shared/schemas/recurringInvoices';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useRecurringInvoices() {
  return useQuery({
    queryKey: ['recurring-invoices'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/finance/recurring-invoices'));
      // Backend ListResponse key is `recurring_invoices` -> camelCased to `recurringInvoices`.
      return z.array(recurringInvoiceSchema).parse(unwrapList(res, 'recurringInvoices'));
    },
  });
}

/**
 * Matches the deployed onion CreateRequest (recurring_invoices handlers.go):
 * { template_invoice_id, frequency, next_invoice_date, end_date, is_active }.
 * Written camelCase here; @voxire/api converts to snake_case on the wire.
 */
interface CreateRecurringPayload {
  templateInvoiceId: number;
  frequency: string;
  nextInvoiceDate: string;
  endDate?: string;
  isActive?: boolean;
}

export function useCreateRecurringInvoice() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateRecurringPayload): Promise<unknown> =>
      unwrap(safe.post('/finance/recurring-invoices', data as unknown as object)),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['recurring-invoices'] });
    },
    onError: () => toast.error('Failed to create recurring invoice'),
  });
}

/**
 * The backend exposes pause/resume as dedicated POST endpoints with the id in
 * the path: POST /finance/recurring-invoices/{id}/pause | /resume.
 */
export function useSetRecurringStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: 'active' | 'paused' }): Promise<void> => {
      if (status === 'paused') {
        return unwrap(safe.post<void>(`/finance/recurring-invoices/${id}/pause`, {}));
      }
      return unwrap(safe.post<void>(`/finance/recurring-invoices/${id}/resume`, {}));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recurring-invoices'] }),
    onError: () => toast.error('Failed to update recurring invoice status'),
  });
}
