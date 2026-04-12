import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi, StorefrontConfig } from '../../shared/services/storeService';
import { storefrontConfigSchema } from '../../shared/schemas/storefront';

export function useStorefrontConfig() {
  return useQuery({
    queryKey: ['store-storefront'],
    queryFn: async () => {
      const res = await unwrap(storeApi.getStorefrontConfig());
      return storefrontConfigSchema.parse(res);
    },
  });
}

export function useUpdateStorefrontConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StorefrontConfig>) => unwrap(storeApi.updateStorefrontConfig(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-storefront'] }),
  });
}
