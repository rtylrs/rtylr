import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useInvoice(id: string | undefined) {
  return useQuery({
    queryKey: ['invoice-detail', id],
    queryFn: () => unwrap(financeApi.getInvoiceDetail(id!)),
    enabled: !!id,
  });
}
