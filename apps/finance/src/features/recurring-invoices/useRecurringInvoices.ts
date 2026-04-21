import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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

// Backend accepts a partial update on /finance/recurring-invoices; we only
// need to flip the status field here. The contract's CreateRecurringInvoiceRequest
// does not model status, so we cast through unknown to keep the call strongly
// typed at every other site that does care about the full shape.
export function useSetRecurringStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'active' | 'paused' }) =>
      unwrap(
        financeApi.updateRecurringInvoice(id, { status } as unknown as Parameters<
          typeof financeApi.updateRecurringInvoice
        >[1])
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recurring-invoices'] }),
  });
}
