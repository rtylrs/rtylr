import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { storeApi, CreateCategoryInput } from '../../shared/services/storeService';
import { storeCategorySchema } from '../../shared/schemas/categories';

export function useCategories() {
  return useQuery({
    queryKey: ['store-categories'],
    queryFn: async () => {
      const res = await unwrap(storeApi.listCategories());
      const raw = (res as any)?.categories ?? res;
      return z.array(storeCategorySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCategoryInput) => unwrap(storeApi.createCategory(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-categories'] }),
  });
}

export function useUpdateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<CreateCategoryInput> & { id: number }) =>
      unwrap(storeApi.updateCategory(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-categories'] }),
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(storeApi.deleteCategory(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-categories'] }),
  });
}
