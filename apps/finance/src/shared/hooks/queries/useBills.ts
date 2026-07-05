import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { ListBillsResponse } from '@voxire/voxie/finance/bill/v1/list_bills';
import type {
  CreateBillRequest,
  CreateBillResponse,
} from '@voxire/voxie/finance/bill/v1/create_bill';
import { safe } from '../../lib/api';

interface ListBillsParams {
  limit: number;
  offset: number;
  status?: string;
  supplierId?: number;
}

export const useListBills = (params: ListBillsParams) => {
  return useQuery<ListBillsResponse>({
    queryKey: ['bills', 'filter', params],
    queryFn: async (): Promise<ListBillsResponse> => {
      const res = await unwrap(safe.get<ListBillsResponse>('/finance/bills', { params }));
      return res as ListBillsResponse;
    },
  });
};

export const useCreateBill = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateBillRequest): Promise<CreateBillResponse> => {
      const res = await unwrap(safe.post<CreateBillResponse>('/finance/bills', data));
      return res as CreateBillResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bills'] });
      toast.success(t('bills.created'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
