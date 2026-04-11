import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { storeApi, StoreSettings } from '../services/storeService';

export function useStoreSettings() {
  return useQuery({
    queryKey: ['store-settings'],
    queryFn: async () => {
      const r = await storeApi.getSettings();
      return r.data ?? null;
    },
  });
}

export function useUpdateStoreSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StoreSettings>) => storeApi.updateSettings(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-settings'] }),
  });
}
