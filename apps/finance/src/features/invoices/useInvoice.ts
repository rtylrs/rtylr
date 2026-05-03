import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import { invoiceSchema } from '../../shared/schemas/invoices';
import type { Invoice } from '../../shared/services/financeApi';

export function useInvoice(id: string | undefined) {
  return useQuery<Invoice>({
    queryKey: ['invoice', id],
    queryFn: async (): Promise<Invoice> => {
      const res = await unwrap(safe.get<unknown>(`/finance/invoices/${id}`));
      return invoiceSchema.parse(res) as Invoice;
    },
    enabled: !!id,
  });
}
