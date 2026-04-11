import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, Category } from '../../shared/services/erpApi';

export function useCategories() {
  return useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await erpApi.listCategories();
      return res.data ?? [];
    },
  });
}
