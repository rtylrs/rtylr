import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { storeApi, CreateShippingZoneInput } from '../../shared/services/storeService';
import { shippingZoneSchema, shippingMethodSchema } from '../../shared/schemas/shipping';

export function useShippingZones() {
  return useQuery({
    queryKey: ['store-shipping-zones'],
    queryFn: async () => {
      const res = await unwrap(storeApi.listShippingZones());
      const raw = (res as any)?.zones ?? res;
      return z.array(shippingZoneSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useShippingMethods() {
  return useQuery({
    queryKey: ['store-shipping-methods'],
    queryFn: async () => {
      const res = await unwrap(storeApi.listShippingMethods());
      const raw = (res as any)?.methods ?? res;
      return z.array(shippingMethodSchema).parse(Array.isArray(raw) ? raw : []);
    },
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
    mutationFn: (data: Partial<CreateShippingZoneInput> & { id: number }) =>
      unwrap(storeApi.updateShippingZone(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-shipping-zones'] }),
  });
}
