import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi, CreateCategoryInput } from '../services/storeService';

export function useCategories() {
  return useQuery({
    queryKey: ['store-categories'],
    queryFn: () => unwrap(storeApi.listCategories()),
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
    mutationFn: (data: Partial<CreateCategoryInput> & { id: number }) => unwrap(storeApi.updateCategory(data)),
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
