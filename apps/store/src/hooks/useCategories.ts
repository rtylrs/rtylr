import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { storeApi, CreateCategoryInput } from '../services/storeService';

export function useCategories() {
  return useQuery({
    queryKey: ['store-categories'],
    queryFn: async () => {
      const r = await storeApi.listCategories();
      return r.data ?? [];
    },
  });
}

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCategoryInput) => storeApi.createCategory(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-categories'] }),
  });
}

export function useUpdateCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<CreateCategoryInput> & { id: number }) => storeApi.updateCategory(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-categories'] }),
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => storeApi.deleteCategory(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-categories'] }),
  });
}
