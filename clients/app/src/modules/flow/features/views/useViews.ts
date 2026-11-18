import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiView } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';
import { PRIORITY_ORDER, type SavedView, type Task } from '../../shared/stores/flowStore';
import { hasActiveTaskFilters, matchesTaskFilters } from '../../shared/filters/taskFilters';
import { viewFiltersToTaskFilters } from '../../shared/filters/viewFilters';
import { todayIso } from '../../shared/lib/dateUtils';

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
    onError: () => toast.error(t('flowErrors.generic')),
  });
}

export function useUpdateView() {
  const qc = useQueryClient();
  return useMutation({
    // The v2 update is a REPLACEMENT: onion#906 made the handler honour every
    // field it receives, and an unsent field arrives as "". So a partial patch
    // would blank whatever it omitted - renaming a view would wipe its filters.
    //
    // Merge the patch over the cached view and send the whole thing. The list
    // query is the same data the caller is looking at, so it is the right
    // baseline; if it is somehow absent we send the patch alone rather than
    // inventing values.
    mutationFn: ({ id, data }: { id: string; data: Partial<SavedView> }): Promise<void> => {
      const cached = qc.getQueryData<SavedView[]>(['views'])?.find((v) => v.id === id);
      const merged: Partial<SavedView> = cached ? { ...cached, ...data } : data;
      return unwrap(flowApi.updateView(id, merged as Partial<ApiView>)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
    onError: () => toast.error(t('flowErrors.generic')),
  });
}

export function useDeleteView() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteView(id));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['views'] }),
    onError: () => toast.error(t('flowErrors.generic')),
  });
}

/**
 * The tasks a view shows, through the same filter model as every other Flow
 * list. `query` is the unsaved search typed on the Views page.
 */
export function useFilteredTasks(
  view: SavedView | null,
  tasks: Task[],
  query = '',
  labelNames?: ReadonlyMap<string, string>
) {
  return useMemo(() => {
    if (!view) return tasks;
    const filters = { ...viewFiltersToTaskFilters(view.filters), query };
    const ctx = { today: todayIso(), labelNames };
    const filtered = hasActiveTaskFilters(filters)
      ? tasks.filter((task) => matchesTaskFilters(task, filters, ctx))
      : [...tasks];
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
  }, [tasks, view, query, labelNames]);
}
