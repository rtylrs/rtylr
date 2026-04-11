import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi, CreateShippingZoneInput } from '../../shared/services/storeService';

export function useShippingZones() {
  return useQuery({
    queryKey: ['store-shipping-zones'],
    queryFn: () => unwrap(storeApi.listShippingZones()),
  });
}

export function useShippingMethods() {
  return useQuery({
    queryKey: ['store-shipping-methods'],
    queryFn: () => unwrap(storeApi.listShippingMethods()),
  });
}

export function useCreateShippingZone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateShippingZoneInput) => unwrap(storeApi.createShippingZone(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-shipping-zones'] }),
  });
}

export function useUpdateShippingZone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<CreateShippingZoneInput> & { id: number }) => unwrap(storeApi.updateShippingZone(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-shipping-zones'] }),
  });
}
