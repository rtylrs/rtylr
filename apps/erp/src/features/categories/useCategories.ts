import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, Category } from '../../shared/services/erpApi';
import { categorySchema } from '../../shared/schemas/categories';

export function useCategories() {
  return useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listCategories());
      return z.array(categorySchema).parse(Array.isArray(res) ? res : []) as Category[];
    },
  });
}
