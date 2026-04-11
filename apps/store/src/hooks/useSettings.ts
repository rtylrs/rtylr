import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi, StoreSettings } from '../services/storeService';

export function useStoreSettings() {
  return useQuery({
    queryKey: ['store-settings'],
    queryFn: () => unwrap(storeApi.getSettings()),
  });
}

export function useUpdateStoreSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StoreSettings>) => unwrap(storeApi.updateSettings(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-settings'] }),
  });
}
