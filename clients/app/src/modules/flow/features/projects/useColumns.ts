import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { unwrap } from '@voxire/api';
import { flowApi, type ApiTask } from '../../shared/services/flowApi';
import { applyBoardMove } from './boardOrder';
import { invalidateTaskProgress } from '../../shared/lib/invalidateTaskProgress';

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
    onSettled: () => invalidateTaskProgress(qc),
  });
}
