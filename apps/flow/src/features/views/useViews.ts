import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { PRIORITY_ORDER, type SavedView, type Task } from '../../shared/stores/flowStore';

export function useViews() {
  return useQuery<SavedView[]>({
    queryKey: ['views'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getViews());
      // backend returns { views: [...] }
      return ((res as any).views ?? (Array.isArray(res) ? res : [])) as SavedView[];
    },
  });
}

export function useCreateView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<SavedView>) => {
      return unwrap(flowApi.createView(data as any));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
  });
}

export function useUpdateView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<SavedView> }) => {
      return unwrap(flowApi.updateView(id, data as any)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
  });
}

export function useDeleteView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteView(id));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
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
      filtered = filtered.filter((t) => t.labels?.some((l) => view.filters.labels!.includes(l)));
    }
    if (view.sortBy) {
      const dir = view.sortDir === 'desc' ? -1 : 1;
      filtered.sort((a, b) => {
        switch (view.sortBy) {
          case 'priority':
            return (PRIORITY_ORDER[a.priority] - PRIORITY_ORDER[b.priority]) * dir;
          case 'dueDate':
            return a.dueDate.localeCompare(b.dueDate) * dir;
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
