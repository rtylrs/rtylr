import React, { useMemo } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { X, Check, Trash2, Archive, Flag } from 'lucide-react';
import { toast } from 'sonner';
import i18next from '../services/i18next';
import { useFlowStore } from '../stores/flowStore';
import {
  useBulkUpdateStatus,
  useBulkUpdatePriority,
  useBulkUpdateArchive,
  useBulkDelete,
  useTasks,
  useTaskDetails,
} from '../../features/tasks/useTasks';
import { useMostRestrictiveProjectRole } from '../lib/currentUser';
import { canEditTask, canDelete } from '../lib/roleUtils';

const t = (key: string, options?: Record<string, unknown>): string =>
  i18next.t(key, options) as string;

export const BulkEditBar: React.FC = () => {
  const { selectedTaskIds, clearSelection } = useFlowStore();
  // Resolve the selection from the task query rather than the Zustand store.
  // `setTasks` has no call site anywhere in the app, so `tasks` was
  // permanently [], no selected task could ever be resolved, and the guard
  // below fell through to "no role" for every user.
  const { data: taskList = [] } = useTasks({});
  // The unfiltered list is a single page (the server caps it), so a task
  // selected on a project board past that page is missing from it. Resolve the
  // rest by id rather than failing the whole selection closed.
  const unresolvedIds = useMemo(
    () => selectedTaskIds.filter((id) => !taskList.some((task) => String(task.id) === String(id))),
    [selectedTaskIds, taskList]
  );
  const taskDetails = useTaskDetails(unresolvedIds);
  const bulkStatus = useBulkUpdateStatus();
  const bulkPriority = useBulkUpdatePriority();
  const bulkArchive = useBulkUpdateArchive();
  const bulkDelete = useBulkDelete();

  // Selection is global and cross-project views (Triage, My issues) let it span
  // projects, so a bulk action can reach several projects at once. Gate on the
  // most restrictive role the user holds across them, and deny outright if any
  // selected task can't be resolved to a project - a missing task must not
  // quietly shrink the set of projects being checked.
  const { projectIds, allResolved } = useMemo(() => {
    const projectIdByTask = new Map<string, string>(
      taskList.map((task): [string, string] => [String(task.id), String(task.projectId)])
    );
    unresolvedIds.forEach((id, index) => {
      const detail = taskDetails[index]?.data;
      if (detail) projectIdByTask.set(String(id), String(detail.task.projectId));
    });

    const ids = new Set<string>();
    let resolved = true;
    for (const taskId of selectedTaskIds) {
      // Ids are compared as strings: the wire sends int64 task and project ids
      // as numbers even where the hand-rolled ApiTask types them as strings.
      const projectId = projectIdByTask.get(String(taskId));
      if (projectId === undefined) {
        resolved = false;
        continue;
      }
      ids.add(projectId);
    }
    return { projectIds: [...ids].sort(), allResolved: resolved };
  }, [selectedTaskIds, taskList, taskDetails, unresolvedIds]);

  const restrictingRole = useMostRestrictiveProjectRole(projectIds);
  const effectiveRole = allResolved ? restrictingRole : undefined;
  const canEdit = effectiveRole !== undefined && canEditTask(effectiveRole);
  const canRemove = effectiveRole !== undefined && canDelete(effectiveRole);
  // Distinguish "you may not" from "we could not work out which projects this
  // selection touches". Blaming view-only access for the latter sent anyone
  // debugging it after the user's role rather than after the data.
  const editBlockedReason = allResolved ? 'flowCommon.readOnlyAccess' : 'flow.bulkEdit.unresolved';
  const deleteBlockedReason = allResolved
    ? 'flowCommon.deleteRestricted'
    : 'flow.bulkEdit.unresolved';

  if (selectedTaskIds.length === 0) return null;

  const isPending =
    bulkStatus.isPending || bulkPriority.isPending || bulkArchive.isPending || bulkDelete.isPending;

  return (
    <AnimatePresence>
      <motion.div
        // `isPending` goes through `animate`, not `style`. A style opacity beside
        // an animated one pins the value and the animation stops running; on
        // TaskCard the same conflict left every board card invisible
        // (rtylr#1078). Here the style always set a number, so the bar was
        // visible and only its entrance was dead, which is why it went unnoticed.
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: isPending ? 0.7 : 1, y: 0 }}
        exit={{ opacity: 0, y: 20 }}
        className="fixed left-1/2 z-50 flex max-w-[calc(100vw-1.5rem)] -translate-x-1/2 items-center gap-1 overflow-x-auto rounded-xl px-3 py-2"
        style={{
          bottom: 'max(1.5rem, env(safe-area-inset-bottom))',
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        <span className="text-flow-meta font-semibold px-2" style={{ color: 'var(--flow-text)' }}>
          {t('flow.bulkEdit.selectedCount', { count: selectedTaskIds.length })}
        </span>
        <div className="h-4 w-px mx-1" style={{ background: 'var(--flow-border)' }} />
        <button
          type="button"
          disabled={isPending || !canEdit}
          title={!canEdit ? t(editBlockedReason) : undefined}
          onClick={() =>
            bulkStatus
              .mutateAsync({ ids: selectedTaskIds, status: 'done' })
              .then(() => clearSelection())
              .catch(() => toast.error(t('flowErrors.generic')))
          }
          className="h-7 px-2 rounded-sm text-flow-meta flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Check size={11} /> {t('flow.bulkEdit.markDone')}
        </button>
        <button
          type="button"
          disabled={isPending || !canEdit}
          title={!canEdit ? t(editBlockedReason) : undefined}
          onClick={() =>
            bulkPriority
              .mutateAsync({ ids: selectedTaskIds, priority: 'urgent' })
              .then(() => clearSelection())
              .catch(() => toast.error(t('flowErrors.generic')))
          }
          className="h-7 px-2 rounded-sm text-flow-meta flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Flag size={11} /> {t('flow.bulkEdit.setUrgent')}
        </button>
        <button
          type="button"
          disabled={isPending || !canEdit}
          title={!canEdit ? t(editBlockedReason) : undefined}
          onClick={() =>
            bulkArchive
              .mutateAsync({ ids: selectedTaskIds })
              .then(() => clearSelection())
              .catch(() => toast.error(t('flowErrors.generic')))
          }
          className="h-7 px-2 rounded-sm text-flow-meta flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Archive size={11} /> {t('flow.bulkEdit.archive')}
        </button>
        <button
          type="button"
          disabled={isPending || !canRemove}
          title={!canRemove ? t(deleteBlockedReason) : undefined}
          onClick={() =>
            bulkDelete
              .mutateAsync(selectedTaskIds)
              .then(() => clearSelection())
              .catch(() => toast.error(t('flowErrors.generic')))
          }
          className="h-7 px-2 rounded-sm text-flow-meta flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-danger-fg)' }}
        >
          <Trash2 size={11} /> {t('flowCommon.delete')}
        </button>
        <div className="h-4 w-px mx-1" style={{ background: 'var(--flow-border)' }} />
        <button
          type="button"
          onClick={clearSelection}
          aria-label={t('flow.bulkEdit.clearSelectionLabel')}
          className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <X size={12} />
        </button>
      </motion.div>
    </AnimatePresence>
  );
};

export default BulkEditBar;
