import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi, StorefrontConfig } from '../../shared/services/storeService';

export function useStorefrontConfig() {
  return useQuery({
    queryKey: ['store-storefront'],
    queryFn: () => unwrap(storeApi.getStorefrontConfig()),
  });
}

export function useUpdateStorefrontConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StorefrontConfig>) => unwrap(storeApi.updateStorefrontConfig(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-storefront'] }),
  });
}
