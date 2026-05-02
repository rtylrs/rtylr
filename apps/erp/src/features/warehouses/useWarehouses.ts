import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, Warehouse } from '../../shared/services/erpApi';
import { warehouseSchema } from '../../shared/schemas/warehouses';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useWarehouses() {
  return useQuery<Warehouse[]>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listWarehouses());
      return z.array(warehouseSchema).parse(unwrapList(res, 'warehouses')) as Warehouse[];
    },
  });
}

export function useCreateWarehouse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Warehouse>) => unwrap(erpApi.createWarehouse(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
  });
}

export function useDeleteWarehouse() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.deleteWarehouse(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['warehouses'] }),
  });
}
