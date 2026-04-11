import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { supportService } from '../services/supportService';

export function useArticles(params?: { search?: string; category?: string }) {
  return useQuery({
    queryKey: ['articles', params],
    queryFn: () => supportService.getArticles(params).then((r) => r.data),
  });
}

export function useArticle(id: number) {
  return useQuery({
    queryKey: ['article', id],
    queryFn: () => supportService.getArticle(id).then((r) => r.data),
    enabled: !!id,
  });
}

export function useArticleCategories() {
  return useQuery({
    queryKey: ['article-categories'],
    queryFn: () => supportService.getArticleCategories().then((r) => r.data),
  });
}

export function useCreateArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { title: string; body: string; category: string }) =>
      supportService.createArticle(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['articles'] });
    },
  });
}

export function useUpdateArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: { title?: string; body?: string; category?: string } }) =>
      supportService.updateArticle(id, data).then((r) => r.data),
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['articles'] });
      qc.invalidateQueries({ queryKey: ['article', vars.id] });
    },
  });
}

export function useDeleteArticle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => supportService.deleteArticle(id).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['articles'] });
    },
  });
}
