import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, Category } from '../../shared/services/erpApi';
import { unwrapList } from '../../shared/lib/unwrapList';

/**
 * Backend category contract (snake_case -> camelCase via @voxire/api):
 * { id, ownerId, name, description, createdAt, updatedAt }.
 * The backend does NOT return parentId / sortOrder / productCount, so this
 * schema only validates what the API actually sends. The shared `Category`
 * interface still requires productCount / sortOrder, so we map sane defaults
 * onto the parsed rows to keep the public type stable.
 */
const categoryRowSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string().optional().nullable(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export function useCategories() {
  return useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listCategories());
      const rows = z.array(categoryRowSchema).parse(unwrapList(res, 'categories'));
      return rows.map(
        (row): Category => ({
          id: row.id,
          name: row.name,
          description: row.description ?? undefined,
          productCount: 0,
          sortOrder: 0,
        })
      );
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
