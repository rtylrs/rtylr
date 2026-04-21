import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';

// No proto type yet — define locally.
export interface ExpiryItem {
  productId: number;
  name: string;
  sku: string;
  batchId: string;
  expiryDate: string;
  quantity: number;
  warehouseName: string;
}

interface ExpiryParams {
  days: number;
}

interface ExpiryItemsResponse {
  items: ExpiryItem[];
}

export const useExpiryItems = (params: ExpiryParams): UseQueryResult<ExpiryItem[]> => {
  return useQuery<ExpiryItem[]>({
    queryKey: ['erp-expiry-items', 'filter', params],
    queryFn: async (): Promise<ExpiryItem[]> => {
      const res = await unwrap(
        safe.get<ExpiryItemsResponse | ExpiryItem[]>('/erp/expiry', { params })
      );
      if (Array.isArray(res)) return res;
      return (res as ExpiryItemsResponse).items ?? [];
    },
  });
};
