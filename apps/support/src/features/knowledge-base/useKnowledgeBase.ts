import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { supportService } from '../../shared/services/supportService';

export function useArticles(params?: { search?: string; category?: string }) {
  return useQuery({
    queryKey: ['articles', params],
    queryFn: () => unwrap(supportService.getArticles(params)),
  });
}

export function useArticle(id: number) {
  return useQuery({
    queryKey: ['article', id],
    queryFn: () => unwrap(supportService.getArticle(id)),
    enabled: !!id,
  });
}

export function useArticleCategories() {
  return useQuery({
    queryKey: ['article-categories'],
    queryFn: () => unwrap(supportService.getArticleCategories()),
  });
}

export function useCreateArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { title: string; body: string; category: string }) =>
      unwrap(supportService.createArticle(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['articles'] });
    },
  });
}

export function useUpdateArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: { title?: string; body?: string; category?: string } }) =>
      unwrap(supportService.updateArticle(id, data)),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['articles'] });
      qc.invalidateQueries({ queryKey: ['article', vars.id] });
    },
  });
}

export function useDeleteArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(supportService.deleteArticle(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['articles'] });
    },
  });
}
