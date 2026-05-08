import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
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

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Category>) => unwrap(erpApi.createCategory(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    onError: () => toast.error('Failed to create category'),
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.deleteCategory(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    onError: () => toast.error('Failed to delete category'),
  });
}

export function useUpdateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: number; name: string; description?: string }) =>
      unwrap(erpApi.updateCategory(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    onError: () => toast.error('Failed to update category'),
  });
}
