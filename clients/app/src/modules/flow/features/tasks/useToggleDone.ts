import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import type { Task } from '../../shared/stores/flowStore';
import { useProjectStatuses, useSetTaskStatus } from '../projects/useStatuses';
import { defaultStatus, firstDoneStatus, statusIdOf } from '../projects/statusUtils';

const UNDO_MS = 5000;

/**
 * The status each task had before a checkbox closed it, so Undo and unticking
 * can put it back. Client-side only: after a reload it falls back to the
 * project's default status.
 */
const previousStatus = new Map<string, string>();

/** Exposed for tests, which share the module between cases. */
export function forgetPreviousStatuses(): void {
  previousStatus.clear();
}

/**
 * One-click done for a list row: ticking moves the task to the project's first
 * done status and offers Undo; unticking (or Undo) restores the status it had.
 */
export function useToggleDone(
  task: Pick<Task, 'id' | 'projectId' | 'title' | 'statusId' | 'sectionId'>
) {
  const { t } = useTranslation();
  const { data: statuses = [] } = useProjectStatuses(task.projectId);
  const setTaskStatus = useSetTaskStatus();

  const reopenTarget = (): string | undefined =>
    previousStatus.get(task.id) ?? defaultStatus(statuses)?.id;

  const reopen = (onError?: () => void): void => {
    const target = reopenTarget();
    if (!target) return;
    previousStatus.delete(task.id);
    setTaskStatus.mutate({ taskId: task.id, statusId: target }, { onError });
  };

  const complete = (onError?: () => void): void => {
    const done = firstDoneStatus(statuses);
    if (!done) return;
    const current = statusIdOf(task);
    if (current && current !== done.id) previousStatus.set(task.id, current);
    setTaskStatus.mutate(
      { taskId: task.id, statusId: done.id },
      {
        onError,
        onSuccess: () => {
          toast.success(t('flow.projectList.markedDone', { title: task.title }), {
            duration: UNDO_MS,
            action: { label: t('flow.projectList.undo'), onClick: () => reopen() },
          });
        },
      }
    );
  };

  return {
    /** False until the project's statuses have loaded. */
    ready: firstDoneStatus(statuses) !== undefined,
    complete,
    reopen,
  };
}
