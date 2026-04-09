import { useQuery } from '@tanstack/react-query';
import { erpApi, Category } from '../services/erpApi';
import { mockCategories } from '../lib/mockData';

export function useCategories() {
  return useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: async () => {
      try {
        const res = await erpApi.listCategories();
        if (res.data && res.data.length) return res.data;
        return mockCategories;
      } catch {
        return mockCategories;
      }
    },
  });
}
