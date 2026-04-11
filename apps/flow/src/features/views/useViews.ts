import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { useFlowStore, PRIORITY_ORDER, type SavedView, type Task } from '../../shared/stores/flowStore';

export function useViews() {
  const views = useFlowStore((s) => s.views);
  const q = useQuery<SavedView[]>({
    queryKey: ['views'],
    queryFn: () => unwrap(flowApi.getViews()) as Promise<SavedView[]>,
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? views : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}

export function useCreateView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<SavedView>) => {
      if (isFlowLocal()) {
        const id = `view-${Date.now()}`;
        const row: SavedView = {
          id,
          name: data.name ?? 'New view',
          icon: data.icon ?? 'eye',
          filters: data.filters ?? {},
          groupBy: data.groupBy,
          sortBy: data.sortBy,
          sortDir: data.sortDir,
        };
        useFlowStore.setState({ views: [...useFlowStore.getState().views, row] });
        return Promise.resolve(row as any);
      }
      return flowApi.createView(data as any);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
  });
}

export function useDeleteView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      if (isFlowLocal()) {
        useFlowStore.setState({
          views: useFlowStore.getState().views.filter((v) => v.id !== id),
        });
        return Promise.resolve();
      }
      return flowApi.deleteView(id);
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
      filtered = filtered.filter((t) =>
        t.labels?.some((l) => view.filters.labels!.includes(l))
      );
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
