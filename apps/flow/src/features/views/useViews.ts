import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiView } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';
import {
  useFlowStore,
  PRIORITY_ORDER,
  type SavedView,
  type Task,
} from '../../shared/stores/flowStore';

const t = (key: string): string => i18next.t(key) as string;

export function useViews() {
  return useQuery<SavedView[]>({
    queryKey: ['views'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getViews());
      return unwrapListResponse<SavedView>(res, 'views');
    },
  });
}

export function useCreateView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<SavedView>) => {
      return unwrap(flowApi.createView(data as Partial<ApiView>));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
    onError: () => toast.error(t('errors.generic')),
  });
}

export function useUpdateView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<SavedView> }): Promise<void> => {
      return unwrap(flowApi.updateView(id, data as Partial<ApiView>)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
    onError: () => toast.error(t('errors.generic')),
  });
}

export function useDeleteView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteView(id));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
    onError: () => toast.error(t('errors.generic')),
  });
}

export function useFilteredTasks(view: SavedView | null, tasks: Task[]) {
  return useMemo(() => {
    if (!view) return tasks;
    let filtered = [...tasks];
    if (view.filters.status && view.filters.status.length) {
      filtered = filtered.filter((t) => view.filters.status!.includes(t.status));
    }
    if (view.filters.priority && view.filters.priority.length) {
      filtered = filtered.filter((t) => view.filters.priority!.includes(t.priority));
    }
    if (view.filters.assigneeId) {
      filtered = filtered.filter((t) => t.assigneeId === view.filters.assigneeId);
    }
    if (view.filters.projectId) {
      filtered = filtered.filter((t) => t.projectId === view.filters.projectId);
    }
    if (view.filters.labels && view.filters.labels.length) {
      filtered = filtered.filter((t) => t.labelIds?.some((l) => view.filters.labels!.includes(l)));
    }
    if (view.sortBy) {
      const dir = view.sortDir === 'desc' ? -1 : 1;
      filtered.sort((a, b) => {
        switch (view.sortBy) {
          case 'priority':
            return (PRIORITY_ORDER[a.priority] - PRIORITY_ORDER[b.priority]) * dir;
          case 'dueDate':
            return (a.dueDate ?? '').localeCompare(b.dueDate ?? '') * dir;
          case 'updated':
            return a.updatedAt.localeCompare(b.updatedAt) * dir;
          case 'created':
            return a.createdAt.localeCompare(b.createdAt) * dir;
          default:
            return 0;
        }
      });
    }
    return filtered;
  }, [tasks, view]);
}
