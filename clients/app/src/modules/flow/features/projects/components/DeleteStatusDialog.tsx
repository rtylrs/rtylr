import React, { useEffect, useId, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AlertTriangle } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../../../../shared/components/ui/dialog';
import { Button } from '../../../../../shared/components/ui/button';
import type { ApiStatus } from '../../../shared/services/flowApi';
import { useTasks } from '../../tasks/useTasks';
import { useDeleteStatus } from '../useStatuses';
import { defaultStatus, displayStatusName, isOnlyDoneStatus, sortStatuses } from '../statusUtils';

interface DeleteStatusDialogProps {
  projectId: string;
  /** The status to delete; the dialog is open while this is set. */
  status: ApiStatus | null;
  statuses: readonly ApiStatus[];
  onClose: () => void;
}

/**
 * Deleting a status that still has tasks asks where they go, because the
 * server refuses the delete without a target. An empty status gets a plain
 * confirm. The project's only done status cannot be deleted at all.
 */
const DeleteStatusDialog = ({
  projectId,
  status,
  statuses,
  onClose,
}: DeleteStatusDialogProps): React.JSX.Element => {
  const { t } = useTranslation();
  const selectId = useId();
  const deleteStatus = useDeleteStatus(projectId);
  // Same query as the board and list, so this is normally already cached.
  const { data: apiTasks = [] } = useTasks({ projectId, activeOnly: true, compact: true });
  const [target, setTarget] = useState('');

  useEffect(() => {
    if (status) setTarget('');
  }, [status]);

  const others = useMemo(
    () => sortStatuses(statuses).filter((s) => s.id !== status?.id),
    [statuses, status]
  );
  // Archived tasks count too: they still sit on the status server-side.
  const taskCount = status
    ? apiTasks.filter((task) => String(task.statusId || task.sectionId || '') === status.id).length
    : 0;
  const protectedDone = status ? isOnlyDoneStatus(status, statuses) : false;
  // The board loads open tasks only, so a done or cancelled status may hold
  // tasks the count cannot see: always ask where they go.
  const countKnown = status ? status.bucket === 'backlog' || status.bucket === 'active' : true;
  const needsTarget = taskCount > 0 || !countKnown;

  const confirm = (): void => {
    if (!status || protectedDone || (needsTarget && !target)) return;
    // An empty status still sends a target: a task the list did not load would
    // otherwise turn the delete into a 409.
    const moveToStatusId = needsTarget ? target : defaultStatus(others)?.id;
    deleteStatus.mutate({ id: status.id, moveToStatusId }, { onSuccess: onClose });
  };

  return (
    <Dialog open={status !== null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="portal-flow-dialog max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle size={16} style={{ color: 'var(--flow-danger-fg)' }} aria-hidden />
            {t('flow.status.deleteTitle', {
              name: status ? displayStatusName(status.name, t) : '',
            })}
          </DialogTitle>
          <DialogDescription>
            {protectedDone
              ? t('flow.status.lastDoneProtected')
              : taskCount > 0
                ? t('flow.status.deleteWithTasks', { count: taskCount })
                : needsTarget
                  ? t('flow.status.deleteChooseTarget')
                  : t('flow.status.deleteEmpty')}
          </DialogDescription>
        </DialogHeader>

        {needsTarget && !protectedDone && (
          <div className="px-6 pb-6">
            <label
              htmlFor={selectId}
              className="mb-1 block text-flow-meta"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.status.moveTasksTo')}
            </label>
            <select
              id={selectId}
              value={target}
              onChange={(e) => setTarget(e.target.value)}
              className="w-full rounded-lg px-3 py-2 text-flow-body"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border-control)',
                color: 'var(--flow-text)',
              }}
            >
              <option value="">{t('flow.status.chooseTarget')}</option>
              {others.map((s) => (
                <option key={s.id} value={s.id}>
                  {displayStatusName(s.name, t)}
                </option>
              ))}
            </select>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {t('flow.status.cancel')}
          </Button>
          {!protectedDone && (
            <Button
              variant="destructive"
              disabled={(needsTarget && !target) || deleteStatus.isPending}
              onClick={confirm}
            >
              {deleteStatus.isPending ? t('flow.confirm.working') : t('flow.status.deleteConfirm')}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default DeleteStatusDialog;
