import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { supportService } from '../../shared/services/supportService';
import { articleSchema, articleCategorySchema } from '../../shared/schemas/knowledgeBase';

export function useArticles(params?: { search?: string; category?: string }) {
  return useQuery({
    queryKey: ['articles', params],
    queryFn: async () => {
      const res = await unwrap(supportService.getArticles(params));
      const raw = res ?? [];
      return z.array(articleSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useArticle(id: number) {
  return useQuery({
    queryKey: ['article', id],
    queryFn: async () => {
      const res = await unwrap(supportService.getArticle(id));
      return articleSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useArticleCategories() {
  return useQuery({
    queryKey: ['article-categories'],
    queryFn: async () => {
      const res = await unwrap(supportService.getArticleCategories());
      const raw = res ?? [];
      return z.array(articleCategorySchema).parse(Array.isArray(raw) ? raw : []);
    },
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
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: { title?: string; body?: string; category?: string };
    }) => unwrap(supportService.updateArticle(id, data)),
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
