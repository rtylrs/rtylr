import React from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { X, Check, Trash2, Archive, Flag } from 'lucide-react';
import { toast } from 'sonner';
import i18next from '../services/i18next';
import { useFlowStore } from '../stores/flowStore';
import {
  useBulkUpdateStatus,
  useBulkUpdatePriority,
  useBulkDelete,
} from '../../features/tasks/useTasks';
import { useCanEditTask, useCanDeleteResource } from '../lib/currentUser';

const t = (key: string): string => i18next.t(key) as string;

export const BulkEditBar: React.FC = () => {
  const { selectedTaskIds, clearSelection } = useFlowStore();
  const bulkStatus = useBulkUpdateStatus();
  const bulkPriority = useBulkUpdatePriority();
  const bulkDelete = useBulkDelete();
  const canEdit = useCanEditTask();
  const canRemove = useCanDeleteResource();

  if (selectedTaskIds.length === 0) return null;

  const isPending = bulkStatus.isPending || bulkPriority.isPending || bulkDelete.isPending;

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
            bulkStatus
              .mutateAsync({ ids: selectedTaskIds, status: 'cancelled' })
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
        <button
          onClick={clearSelection}
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
