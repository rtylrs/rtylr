import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ProductServiceClient } from '@voxire/voxie/erp/product/v1/client';
import type { ReorderSuggestion } from '@voxire/voxie/erp/product/v1/types';
import { withErpListBounds } from '../../shared/lib/erpListBounds';

const productClient = createOneKitClient(ProductServiceClient, 'erp');

export type { ReorderSuggestion };

export function useReorderSuggestions() {
  return useQuery<ReorderSuggestion[]>({
    queryKey: ['reorder-suggestions'],
    queryFn: async () => {
      const response = await productClient.listReorderSuggestions(withErpListBounds({}));
      return response.suggestions ?? [];
    },
  });
}
