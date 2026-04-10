import { useQuery } from '@tanstack/react-query';
import { erpApi, Warehouse } from '../services/erpApi';

export function useWarehouses() {
  return useQuery<Warehouse[]>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      const res = await erpApi.listWarehouses();
      return res.data ?? [];
    },
  });
}
