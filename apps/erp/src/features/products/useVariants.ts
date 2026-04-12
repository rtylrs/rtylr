import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ProductVariant } from '../../shared/services/erpApi';
import { productVariantSchema } from '../../shared/schemas/products';

export function useVariants(productId?: number) {
  return useQuery<ProductVariant[]>({
    queryKey: ['variants', productId],
    queryFn: async () => {
      const res = await unwrap(erpApi.listVariants(productId));
      return z.array(productVariantSchema).parse(Array.isArray(res) ? res : []) as ProductVariant[];
    },
  });
}
