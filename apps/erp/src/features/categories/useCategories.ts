import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, Category } from '../../shared/services/erpApi';
import { categorySchema } from '../../shared/schemas/categories';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useCategories() {
  return useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listCategories());
      return z.array(categorySchema).parse(unwrapList(res, 'categories')) as Category[];
    },
  });
}
