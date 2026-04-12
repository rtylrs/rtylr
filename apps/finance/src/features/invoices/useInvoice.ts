import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi, Invoice } from '../../shared/services/financeApi';
import { invoiceSchema } from '../../shared/schemas/invoices';

export function useInvoice(id: string | undefined) {
  return useQuery({
    queryKey: ['invoice-detail', id],
    queryFn: async () => {
      const res = await unwrap(financeApi.getInvoiceDetail(id!));
      return invoiceSchema.parse(res) as Invoice;
    },
    enabled: !!id,
  });
}
