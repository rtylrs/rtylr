import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useInvoice(id: string | undefined) {
  return useQuery({
    queryKey: ['invoice-detail', id],
    queryFn: () => financeApi.getInvoiceDetail(id!),
    enabled: !!id,
  });
}
