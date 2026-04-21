import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';

// No proto type yet — define locally. See docs/ERP-CONTRACTS.md follow-up.
export interface ProductVariant {
  id: number;
  productId: number;
  name: string;
  sku: string;
  attributes: Record<string, string>;
  stock: number;
  price: number;
}

interface ListVariantsResponse {
  variants: ProductVariant[];
}

export const useListVariants = (productId?: number): UseQueryResult<ProductVariant[]> => {
  const params = productId !== undefined ? { productId } : {};
  return useQuery<ProductVariant[]>({
    queryKey: ['erp-variants', 'filter', params],
    queryFn: async (): Promise<ProductVariant[]> => {
      const res = await unwrap(
        safe.get<ListVariantsResponse>('/erp/variants', {
          params: productId !== undefined ? { productId } : undefined,
        })
      );
      return (res as ListVariantsResponse).variants ?? [];
    },
  });
};
