import { useQuery } from '@tanstack/react-query';
import { erpApi, Brand } from '../services/erpApi';
import { mockBrands } from '../lib/mockData';

export function useBrands() {
  return useQuery<Brand[]>({
    queryKey: ['brands'],
    queryFn: async () => {
      try {
        const res = await erpApi.listBrands();
        if (res.data && res.data.length) return res.data;
        return mockBrands;
      } catch {
        return mockBrands;
      }
    },
  });
}
