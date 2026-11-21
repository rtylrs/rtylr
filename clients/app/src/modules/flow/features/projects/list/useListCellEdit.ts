import { useQueryClient, type QueryClient } from '@tanstack/react-query';
import type {
  ApiStatus,
  ApiSubtask,
  ApiTask,
  ApiTaskDetail,
} from '../../../shared/services/flowApi';
import { useUpdateTask } from '../../tasks/useTasks';
import { useSetTaskLabels } from '../../tasks/useLabels';
import { useSetTaskStatus } from '../useStatuses';

const sameId = (id: unknown, taskId: string): boolean => String(id) === taskId;

/**
 * Writes a change into every cached copy of a task before the server answers:
 * the task lists (`['tasks', ...]`) and the detail queries, where an expanded
 * row's subtasks come from (`['task', parentId]`). The mutation's own success
 * handler then invalidates both, so the server's values replace these.
 */
export function patchTaskInCaches(qc: QueryClient, taskId: string, patch: Partial<ApiTask>): void {
  qc.setQueriesData<ApiTask[]>({ queryKey: ['tasks'] }, (old) =>
    Array.isArray(old)
      ? old.map((task) => (sameId(task.id, taskId) ? { ...task, ...patch } : task))
      : old
  );
  qc.setQueriesData<ApiTaskDetail>({ queryKey: ['task'] }, (old) => {
    if (!old || typeof old !== 'object' || !('task' in old)) return old;
    const own = !!old.task && sameId(old.task.id, taskId);
    const subtasks = Array.isArray(old.subtasks) ? old.subtasks : undefined;
    const inSubtasks = !!subtasks?.some((subtask) => sameId(subtask.id, taskId));
    if (!own && !inSubtasks) return old;
    return {
      ...old,
      task: own ? { ...old.task, ...patch } : old.task,
      subtasks:
        inSubtasks && subtasks
          ? subtasks.map((subtask) =>
              sameId(subtask.id, taskId) ? ({ ...subtask, ...patch } as ApiSubtask) : subtask
            )
          : old.subtasks,
    };
  });
}

/**
 * In-place edits from the List's cells, through the same mutations the task
 * panel uses (and so the same toasts and invalidation). The cell shows the new
 * value at once; a failed write refetches, which puts the old value back.
 *
 * The editor that calls this unmounts as soon as the value is chosen, and a
 * per-call `onError` never runs once its observer has gone. The promise from
 * `mutateAsync` settles regardless, so the rollback hangs off that instead.
 */
export function useListCellEdit() {
  const qc = useQueryClient();
  const updateTask = useUpdateTask();
  const setTaskStatus = useSetTaskStatus();
  const setTaskLabels = useSetTaskLabels();

  const optimistic = (taskId: string, patch: Partial<ApiTask>): void => {
    // A list refetch already in flight would land with the old value.
    void qc.cancelQueries({
      queryKey: ['tasks'],
      predicate: (query) => query.state.data !== undefined,
    });
    patchTaskInCaches(qc, taskId, patch);
  };
  const rollback = (): void => {
    void qc.invalidateQueries({ queryKey: ['tasks'] });
    void qc.invalidateQueries({ queryKey: ['task'] });
  };

  return {
    /**
     * A partial update. `display` holds fields the server derives (a name for
     * an assignee id, say) so the cell can show them before it answers.
     */
    update(taskId: string, data: Partial<ApiTask>, display: Partial<ApiTask> = {}): void {
      optimistic(taskId, { ...data, ...display });
      updateTask.mutateAsync({ id: taskId, data }).catch(rollback);
    },
    setStatus(taskId: string, status: ApiStatus): void {
      optimistic(taskId, {
        statusId: status.id,
        sectionId: status.id,
        statusName: status.name,
        statusColor: status.color,
        status: status.legacyStatus as ApiTask['status'],
      });
      setTaskStatus.mutateAsync({ taskId, statusId: status.id }).catch(rollback);
    },
    setLabels(taskId: string, labelIds: string[]): void {
      optimistic(taskId, { labels: labelIds });
      setTaskLabels
        .mutateAsync({ task_id: Number(taskId), label_ids: labelIds.map(Number) })
        .catch(rollback);
    },
  };
}

export type ListCellEdit = ReturnType<typeof useListCellEdit>;
