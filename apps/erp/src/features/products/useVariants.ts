import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ProductVariant } from '../../shared/services/erpApi';

export function useVariants(productId?: number) {
  return useQuery<ProductVariant[]>({
    queryKey: ['variants', productId],
    queryFn: async () => {
      const res = await erpApi.listVariants(productId);
      return res.data ?? [];
    },
  });
}
