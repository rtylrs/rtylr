import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { storeApi, StorefrontConfig } from '../services/storeService';

export function useStorefrontConfig() {
  return useQuery({
    queryKey: ['store-storefront'],
    queryFn: async () => {
      const r = await storeApi.getStorefrontConfig();
      return r.data ?? null;
    },
  });
}

export function useUpdateStorefrontConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StorefrontConfig>) => storeApi.updateStorefrontConfig(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-storefront'] }),
  });
}
