import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';
import {
  flowApi,
  type ApiRelationDirection,
  type ApiRelationType,
  type ApiTaskRelationGroup,
} from '../../shared/services/flowApi';

const t = (key: string): string => i18next.t(key) as string;

export const taskRelationsKey = (taskId: string) => ['task-relations', taskId] as const;

/**
 * A task's relations, grouped by type and direction. `enabled` lets the
 * blocked badge fetch them only once someone points at it.
 */
export function useTaskRelations(taskId: string | undefined, enabled = true) {
  return useQuery<ApiTaskRelationGroup[]>({
    queryKey: taskRelationsKey(taskId ?? ''),
    queryFn: () => unwrap(flowApi.listTaskRelations(taskId!)),
    enabled: !!taskId && enabled,
  });
}

// A relation shows on both tasks, and a blocks relation changes the blocked
// badge wherever either task is listed.
function invalidateBothEnds(qc: QueryClient, taskIds: string[]) {
  for (const id of taskIds) {
    void qc.invalidateQueries({ queryKey: taskRelationsKey(id) });
    void qc.invalidateQueries({ queryKey: ['task', id] });
  }
  void qc.invalidateQueries({ queryKey: ['tasks'] });
  void qc.invalidateQueries({ queryKey: ['my-tasks'] });
}

/**
 * A refused relation is 409 when the tasks are already linked that way or a
 * blocks link would close a loop, and 400 for a task linked to itself.
 */
function relationErrorMessage(error: unknown): string {
  const status = (error as { status?: unknown } | null)?.status;
  if (status === 409) return t('flow.taskDetail.relationConflict');
  if (status === 400) return t('flow.taskDetail.relationSelf');
  return t('flowToasts.task.updateFailed');
}

export function useAddTaskRelation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: {
      taskId: string;
      otherTaskId: string;
      type: ApiRelationType;
      direction: ApiRelationDirection;
    }) =>
      unwrap(flowApi.addTaskRelation(input.taskId, input.otherTaskId, input.type, input.direction)),
    onSuccess: (_data, { taskId, otherTaskId }) => invalidateBothEnds(qc, [taskId, otherTaskId]),
    onError: (error) => toast.error(relationErrorMessage(error)),
  });
}

export function useRemoveTaskRelation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { taskId: string; relationId: string; otherTaskId: string }) =>
      unwrap(flowApi.removeTaskRelation(input.taskId, input.relationId)),
    onSuccess: (_data, { taskId, otherTaskId }) => invalidateBothEnds(qc, [taskId, otherTaskId]),
    onError: () => toast.error(t('flowToasts.task.updateFailed')),
  });
}
