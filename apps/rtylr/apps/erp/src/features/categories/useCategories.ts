import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { CategoryServiceClient } from '@voxire/voxie/erp/category/v1/client';
import type {
  Category,
  CreateCategoryRequest,
  CreateCategoryResponse,
  DeleteCategoryResponse,
  UpdateCategoryRequest,
  UpdateCategoryResponse,
} from '@voxire/voxie/erp/category/v1/types';
import { ERP_LIST_LIMIT } from '../../shared/lib/erpListBounds';

const categoryClient = createOneKitClient(CategoryServiceClient, 'erp');

export type { Category };

export async function fetchAllCategories(): Promise<Category[]> {
  const categories: Category[] = [];
  let offset = 0;

  while (true) {
    const response = await categoryClient.listCategories({ limit: ERP_LIST_LIMIT, offset });
    const page = response.categories ?? [];
    categories.push(...page);

    const hasMore =
      response.hasMore ??
      (response.total !== undefined
        ? categories.length < response.total
        : page.length === ERP_LIST_LIMIT);
    if (!hasMore || page.length === 0) return categories;
    offset += page.length;
  }
}

export function useCategories() {
  return useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: fetchAllCategories,
  });
}

export function useCreateCategory() {
  const qc = useQueryClient();
  return useMutation<CreateCategoryResponse, Error, CreateCategoryRequest>({
    mutationFn: (data) => categoryClient.createCategory(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    onError: () => toast.error('Failed to create category'),
  });
}

export function useDeleteCategory() {
  const qc = useQueryClient();
  return useMutation<DeleteCategoryResponse, Error, string>({
    mutationFn: (id) => categoryClient.deleteCategory({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    onError: () => toast.error('Failed to delete category'),
  });
}

export function useUpdateCategory() {
  const qc = useQueryClient();
  return useMutation<UpdateCategoryResponse, Error, UpdateCategoryRequest>({
    mutationFn: (data) => categoryClient.updateCategory(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    onError: () => toast.error('Failed to update category'),
  });
}
