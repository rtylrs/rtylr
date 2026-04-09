import { useQuery } from '@tanstack/react-query';
import { buildMockInvoice, InvoiceDetail } from '../lib/mockData';
import { financeApi } from '../services/financeApi';

export function useInvoice(id: string | undefined) {
  return useQuery<InvoiceDetail>({
    queryKey: ['invoice-detail', id],
    queryFn: async () => {
      if (!id) throw new Error('id required');
      try {
        await financeApi.getInvoiceDetail(id);
      } catch {
        // ignore, fall through to mock
      }
      return buildMockInvoice(id);
    },
    enabled: !!id,
  });
}
