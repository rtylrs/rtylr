import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { storeApi, CreateShippingZoneInput } from '../services/storeService';

export function useShippingZones() {
  return useQuery({
    queryKey: ['store-shipping-zones'],
    queryFn: async () => {
      const r = await storeApi.listShippingZones();
      return r.data ?? [];
    },
  });
}

export function useShippingMethods() {
  return useQuery({
    queryKey: ['store-shipping-methods'],
    queryFn: async () => {
      const r = await storeApi.listShippingMethods();
      return r.data ?? [];
    },
  });
}

export function useCreateShippingZone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateShippingZoneInput) => storeApi.createShippingZone(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-shipping-zones'] }),
  });
}

export function useUpdateShippingZone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<CreateShippingZoneInput> & { id: number }) => storeApi.updateShippingZone(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-shipping-zones'] }),
  });
}
