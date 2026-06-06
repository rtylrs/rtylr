import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import { unwrapList } from '../../shared/lib/unwrapList';
import {
  warehouseSchema,
  Warehouse,
  CreateWarehouseInput,
  UpdateWarehouseInput,
} from './warehouseSchema';

export function useWarehouses() {
  return useQuery<Warehouse[]>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      const res = await unwrap(safe.get<unknown>('/erp/warehouses'));
      return z.array(warehouseSchema).parse(unwrapList(res, 'warehouses'));
    },
  });
}

export function useCreateWarehouse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateWarehouseInput) => unwrap(safe.post<unknown>('/erp/warehouses', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error('Failed to create warehouse'),
  });
}

export function useUpdateWarehouse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateWarehouseInput) =>
      unwrap(safe.put<unknown>(`/erp/warehouses/${data.id}`, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error('Failed to update warehouse'),
  });
}

export function useDeleteWarehouse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(safe.delete<unknown>(`/erp/warehouses/${id}`)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
    onError: () => toast.error('Failed to delete warehouse'),
  });
}
