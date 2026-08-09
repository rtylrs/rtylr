import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ProductServiceClient } from '@voxire/voxie/erp/product/v1/client';
import type { ExpiringItem } from '@voxire/voxie/erp/product/v1/types';

const productClient = createOneKitClient(ProductServiceClient, 'erp');

export type ExpiryItem = ExpiringItem;

export function useExpiry(daysAhead = 30) {
  return useQuery<ExpiringItem[]>({
    queryKey: ['expiry', daysAhead],
    queryFn: async () => {
      const response = await productClient.listExpiringItems({ daysAhead });
      return response.items ?? [];
    },
  });
}
