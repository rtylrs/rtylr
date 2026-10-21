import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { PriceListServiceClient } from '@voxire/voxie/erp/price_list/v1/client';
import type {
  CreatePriceListRequest,
  CreatePriceListResponse,
  PriceList,
} from '@voxire/voxie/erp/price_list/v1/types';
import { withErpListBounds } from '../../shared/lib/erpListBounds';

const priceListClient = createOneKitClient(PriceListServiceClient, 'erp');

export function usePriceLists() {
  return useQuery<PriceList[]>({
    queryKey: ['price-lists'],
    queryFn: async () => {
      const response = await priceListClient.listPriceLists(withErpListBounds({}));
      return response.priceLists ?? [];
    },
  });
}

export function useCreatePriceList() {
  const qc = useQueryClient();
  return useMutation<CreatePriceListResponse, Error, CreatePriceListRequest>({
    mutationFn: (data) => priceListClient.createPriceList(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['price-lists'] }),
    onError: () => toast.error('Failed to create price list'),
  });
}
