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
} from '../../features/tasks/useTasks';
import { useMostRestrictiveProjectRole } from '../lib/currentUser';
import { canEditTask, canDelete } from '../lib/roleUtils';

const t = (key: string): string => i18next.t(key) as string;

export const BulkEditBar: React.FC = () => {
  const { selectedTaskIds, clearSelection, tasks } = useFlowStore();
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
    const ids = new Set<string>();
    let resolved = true;
    for (const taskId of selectedTaskIds) {
      const task = tasks.find((candidate) => candidate.id === taskId);
      if (!task) {
        resolved = false;
        continue;
      }
      ids.add(task.projectId);
    }
    return { projectIds: [...ids].sort(), allResolved: resolved };
  }, [selectedTaskIds, tasks]);

  const restrictingRole = useMostRestrictiveProjectRole(projectIds);
  const effectiveRole = allResolved ? restrictingRole : undefined;
  const canEdit = effectiveRole !== undefined && canEditTask(effectiveRole);
  const canRemove = effectiveRole !== undefined && canDelete(effectiveRole);

  if (selectedTaskIds.length === 0) return null;

  const isPending =
    bulkStatus.isPending || bulkPriority.isPending || bulkArchive.isPending || bulkDelete.isPending;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: 20 }}
        className="fixed left-1/2 z-50 flex max-w-[calc(100vw-1.5rem)] -translate-x-1/2 items-center gap-1 overflow-x-auto rounded-xl px-3 py-2"
        style={{
          bottom: 'max(1.5rem, env(safe-area-inset-bottom))',
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
          opacity: isPending ? 0.7 : 1,
        }}
      >
        <span className="text-[11px] font-semibold px-2" style={{ color: 'var(--flow-text)' }}>
          {selectedTaskIds.length} selected
        </span>
        <div className="h-4 w-px mx-1" style={{ background: 'var(--flow-border)' }} />
        <button
          disabled={isPending || !canEdit}
          title={!canEdit ? t('common.readOnlyAccess') : undefined}
          onClick={() =>
            bulkStatus
              .mutateAsync({ ids: selectedTaskIds, status: 'done' })
              .then(() => clearSelection())
              .catch(() => toast.error(t('errors.generic')))
          }
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Check size={11} /> Mark done
        </button>
        <button
          disabled={isPending || !canEdit}
          title={!canEdit ? t('common.readOnlyAccess') : undefined}
          onClick={() =>
            bulkPriority
              .mutateAsync({ ids: selectedTaskIds, priority: 'urgent' })
              .then(() => clearSelection())
              .catch(() => toast.error(t('errors.generic')))
          }
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Flag size={11} /> Set urgent
        </button>
        <button
          disabled={isPending || !canEdit}
          title={!canEdit ? t('common.readOnlyAccess') : undefined}
          onClick={() =>
            bulkArchive
              .mutateAsync({ ids: selectedTaskIds })
              .then(() => clearSelection())
              .catch(() => toast.error(t('errors.generic')))
          }
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Archive size={11} /> Archive
        </button>
        <button
          disabled={isPending || !canRemove}
          title={!canRemove ? t('common.deleteRestricted') : undefined}
          onClick={() =>
            bulkDelete
              .mutateAsync(selectedTaskIds)
              .then(() => clearSelection())
              .catch(() => toast.error(t('errors.generic')))
          }
          className="h-7 px-2 rounded text-[11px] flex items-center gap-1 hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: '#ef4444' }}
        >
          <Trash2 size={11} /> Delete
        </button>
        <div className="h-4 w-px mx-1" style={{ background: 'var(--flow-border)' }} />
        <button type="button"
          onClick={clearSelection}
          aria-label={t('flow.bulkEdit.clearSelectionLabel')}
          className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <X size={12} />
        </button>
      </motion.div>
    </AnimatePresence>
  );
};

export default BulkEditBar;
