import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

// No proto type yet - define locally.
export interface Category {
  id: number;
  name: string;
  description?: string;
  parentId?: number | null;
  productCount: number;
  sortOrder: number;
}

export interface CreateCategoryInput {
  name: string;
  description?: string;
  parentId?: number | null;
  sortOrder?: number;
}

interface ListCategoriesResponse {
  categories: Category[];
}

export const useListCategories = (): UseQueryResult<Category[]> => {
  return useQuery<Category[]>({
    queryKey: ['erp-categories'],
    queryFn: async (): Promise<Category[]> => {
      const res = await unwrap(safe.get<ListCategoriesResponse>('/erp/categories'));
      return (res as ListCategoriesResponse).categories ?? [];
    },
  });
};

export const useCreateCategory = (): UseMutationResult<Category, Error, CreateCategoryInput> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateCategoryInput): Promise<Category> => {
      const res = await unwrap(
        safe.post<Category>('/erp/categories', {
          name: data.name,
          description: data.description,
        })
      );
      return res as Category;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-categories'] });
      toast.success(t('erp.categoryCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
