import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, Warehouse } from '../../shared/services/erpApi';
import { warehouseSchema } from '../../shared/schemas/warehouses';

export function useWarehouses() {
  return useQuery<Warehouse[]>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listWarehouses());
      return z.array(warehouseSchema).parse(Array.isArray(res) ? res : []) as Warehouse[];
    },
  });
}
