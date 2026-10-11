import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { unwrap } from '@voxire/api';
import { flowApi, ApiProjectColumn, type ApiTask } from '../../shared/services/flowApi';
import { applyBoardMove } from './boardOrder';

function key(projectId: string) {
  return ['columns', projectId];
}

const COLUMNS_PAGE_SIZE = 500;

/** The board needs every column, not only the server's first page. */
export async function fetchAllProjectColumns(projectId: string): Promise<ApiProjectColumn[]> {
  const all: ApiProjectColumn[] = [];
  let page = 1;

  for (;;) {
    const res = await unwrap(
      flowApi.getProjectColumns(projectId, { page, limit: COLUMNS_PAGE_SIZE })
    );
    const rows = res.columns ?? [];
    all.push(...rows);

    const pagination = res.pagination;
    const hasNext =
      pagination?.hasNext ??
      (typeof pagination?.total === 'number' && pagination.total > all.length);
    if (!hasNext || rows.length === 0) return all;

    page = (pagination?.page ?? page) + 1;
  }
}

export function useProjectColumns(projectId: string) {
  return useQuery<ApiProjectColumn[]>({
    queryKey: key(projectId),
    queryFn: () => fetchAllProjectColumns(projectId),
    enabled: !!projectId,
  });
}

export function useCreateColumn(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color?: string; sortOrder?: number }) =>
      unwrap(
        flowApi.createProjectColumn({
          project_id: projectId,
          name: data.name,
          color: data.color,
          sort_order: data.sortOrder,
        })
      ).then((r) => r.column),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(projectId) }),
    onError: () => toast.error(t('flowToasts.column.createFailed')),
  });
}

export function useUpdateColumn(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: string; name: string; color?: string; sortOrder?: number }) =>
      unwrap(
        flowApi.updateProjectColumn({
          id: data.id,
          name: data.name,
          color: data.color,
          sort_order: data.sortOrder,
        })
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(projectId) }),
    onError: () => toast.error(t('flowToasts.column.updateFailed')),
  });
}

export function useDeleteColumn(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.deleteProjectColumn(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: key(projectId) });
      qc.invalidateQueries({ queryKey: ['tasks'] });
    },
    onError: () => toast.error(t('flowToasts.column.deleteFailed')),
  });
}

export interface MoveTaskVars {
  taskId: string;
  columnId: string;
  /** The destination column's cards in their new order, the moved card included. */
  orderedTaskIds?: readonly string[];
}

/**
 * Moves a card to a column and, when `orderedTaskIds` is given, stores its place
 * among the others.
 *
 * The cache is updated before the request leaves. The board shows a live
 * preview while a card is dragged and drops that preview on release; if the
 * cards then waited for the server they would snap back to where they were for
 * a moment before jumping to the new place. A failed move restores the cache.
 */
export function useMoveTaskToColumn() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, columnId, orderedTaskIds }: MoveTaskVars) =>
      unwrap(flowApi.moveTaskToColumn(taskId, columnId, orderedTaskIds)),
    onMutate: (vars: MoveTaskVars) => {
      const snapshot = qc.getQueriesData<ApiTask[]>({ queryKey: ['tasks'] });
      qc.setQueriesData<ApiTask[]>({ queryKey: ['tasks'] }, (cached) =>
        cached ? applyBoardMove(cached, vars) : cached
      );
      void qc.cancelQueries({ queryKey: ['tasks'] });
      return { snapshot };
    },
    onError: (_error, _vars, context) => {
      context?.snapshot.forEach(([queryKey, data]) => qc.setQueryData(queryKey, data));
      toast.error(t('flowToasts.task.moveFailed'));
    },
    onSettled: (_data, _error, { taskId }) => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['task', taskId] });
    },
  });
}
