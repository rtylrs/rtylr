import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import {
  ProductServiceClient,
  type ReorderSuggestion,
} from '@voxire/voxie/erp/product/v1/service_client';

const productClient = createOneKitClient(ProductServiceClient, 'erp');

export type { ReorderSuggestion };

export function useReorderSuggestions() {
  return useQuery<ReorderSuggestion[]>({
    queryKey: ['reorder-suggestions'],
    queryFn: async () => {
      const response = await productClient.listReorderSuggestions({});
      return response.suggestions;
    },
  });
}
