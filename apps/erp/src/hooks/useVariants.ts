import { useQuery } from '@tanstack/react-query';
import { erpApi, ProductVariant } from '../services/erpApi';
import { mockVariants } from '../lib/mockData';

export function useVariants(productId?: number) {
  return useQuery<ProductVariant[]>({
    queryKey: ['variants', productId],
    queryFn: async () => {
      try {
        const res = await erpApi.listVariants(productId);
        if (res.data && res.data.length) return res.data;
        return productId ? mockVariants.filter((v) => v.productId === productId) : mockVariants;
      } catch {
        return productId ? mockVariants.filter((v) => v.productId === productId) : mockVariants;
      }
    },
  });
}
