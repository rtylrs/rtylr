import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, ProductVariant } from '../../shared/services/erpApi';
import { productVariantSchema } from '../../shared/schemas/products';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useVariants(productId?: number) {
  return useQuery<ProductVariant[]>({
    queryKey: ['variants', productId],
    queryFn: async () => {
      const res = await unwrap(erpApi.listVariants(productId));
      return z.array(productVariantSchema).parse(unwrapList(res, 'variants')) as ProductVariant[];
    },
  });
}

export function useCreateVariant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ProductVariant>) => unwrap(erpApi.createVariant(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['variants'] }),
    onError: () => toast.error('Failed to create variant'),
  });
}
