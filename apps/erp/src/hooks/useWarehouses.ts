import { useQuery } from '@tanstack/react-query';
import { erpApi, Warehouse } from '../services/erpApi';
import { mockWarehouses } from '../lib/mockData';

export function useWarehouses() {
  return useQuery<Warehouse[]>({
    queryKey: ['warehouses'],
    queryFn: async () => {
      try {
        const res = await erpApi.listWarehouses();
        if (res.data && res.data.length) return res.data;
        return mockWarehouses;
      } catch {
        return mockWarehouses;
      }
    },
  });
}
