import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import {
  Eye,
  Plus,
  Flame,
  Timer,
  CheckCircle2,
  Pencil,
  Trash2,
  Check,
  X,
  Lock,
  Users,
} from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useFlowStore, SavedView } from '../../shared/stores/flowStore';
import {
  useViews,
  useCreateView,
  useUpdateView,
  useDeleteView,
  useFilteredTasks,
} from './useViews';
import TaskRow from '../../shared/components/TaskRow';
import ConfirmDialog from '../../shared/components/ConfirmDialog';

const ICON_MAP: Record<string, React.ComponentType<{ size?: number }>> = {
  flame: Flame,
  timer: Timer,
  'check-circle': CheckCircle2,
};

const Views: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: views = [] } = useViews();
  const createView = useCreateView();
  const updateView = useUpdateView();
  const deleteView = useDeleteView();
  const { tasks: allTasks } = useFlowStore();
  const [selected, setSelected] = useState<SavedView | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [pendingDelete, setPendingDelete] = useState<SavedView | null>(null);

  const displayView = selected || views[0] || null;
  const tasks = useFilteredTasks(displayView, allTasks);

  const startEdit = (v: SavedView, e: React.MouseEvent) => {
    e.stopPropagation();
    setEditingId(v.id);
    setEditName(v.name);
  };

  const commitEdit = (v: SavedView) => {
    if (!editName.trim() || editName === v.name) {
      setEditingId(null);
      return;
    }
    updateView.mutate(
      { id: v.id, data: { name: editName.trim() } },
      {
        onSuccess: () => {
          toast.success(t('toasts.view.renamed'));
          setEditingId(null);
        },
        onError: () => toast.error(t('toasts.view.renameFailed')),
      }
    );
  };

  // An absent visibility means shared: that is the column default and how every
  // view behaved before the field existed, so an older response must not read as
  // private.
  const isPrivate = (v: SavedView): boolean => v.visibility === 'private';

  const toggleVisibility = (v: SavedView, e: React.MouseEvent): void => {
    e.stopPropagation();
    const next = isPrivate(v) ? 'shared' : 'private';
    updateView.mutate(
      { id: v.id, data: { visibility: next } },
      {
        onSuccess: () =>
          toast.success(
            next === 'private' ? t('toasts.view.madePrivate') : t('toasts.view.madeShared')
          ),
        onError: () => toast.error(t('toasts.view.visibilityFailed')),
      }
    );
  };

  const handleDelete = (v: SavedView, e: React.MouseEvent) => {
    e.stopPropagation();
    setPendingDelete(v);
  };

  const handleCreateView = (): void => {
    createView.mutate(
      { name: t('flow.views.defaultName'), icon: 'eye', filters: {} },
      { onSuccess: () => toast.success(t('toasts.view.created')) }
    );
  };

  const performDelete = (v: SavedView) => {
    deleteView.mutate(v.id, {
      onSuccess: () => {
        toast.success(t('toasts.view.deleted'));
        if (selected?.id === v.id) setSelected(null);
        setPendingDelete(null);
      },
      onError: () => toast.error(t('toasts.view.deleteFailed')),
    });
  };

  return (
    <div className="flex h-full">
      <HelmetContainer title={`${t('flow.views.heading')} · Flow`} />
      <aside
        className="w-[256px] shrink-0 px-3 py-4"
        style={{
          background: 'var(--flow-surface)',
          borderRight: '1px solid var(--flow-border)',
        }}
      >
        <div className="flex items-center justify-between px-2 mb-3">
          <h2
            className="text-meta font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.views.heading')}
          </h2>
          <button
            type="button"
            onClick={handleCreateView}
            disabled={createView.isPending}
            className="h-5 w-5 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
            title={t('flow.views.addTooltip')}
            aria-label={t('flow.views.newViewLabel')}
          >
            <Plus size={11} />
          </button>
        </div>

        <div className="space-y-0.5">
          {views.map((v) => {
            const Icon = ICON_MAP[v.icon] || Eye;
            const active = (selected?.id ?? views[0]?.id) === v.id;
            const isEditing = editingId === v.id;

            return (
              <div
                key={v.id}
                className="group flex items-center gap-1 rounded-sm"
                style={{ background: active ? 'var(--flow-active)' : 'transparent' }}
              >
                {isEditing ? (
                  <div className="flex items-center gap-1 flex-1 px-2 py-1">
                    <input
                      autoFocus
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') commitEdit(v);
                        if (e.key === 'Escape') setEditingId(null);
                      }}
                      className="flex-1 bg-transparent text-meta"
                      style={{ color: 'var(--flow-text)' }}
                    />
                    <button
                      type="button"
                      onClick={() => commitEdit(v)}
                      aria-label={t('flow.views.confirmRenameLabel')}
                      className="h-4 w-4 flex items-center justify-center rounded-sm hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--accent)' }}
                    >
                      <Check size={10} />
                    </button>
                    <button
                      type="button"
                      onClick={() => setEditingId(null)}
                      aria-label={t('flow.views.cancelRenameLabel')}
                      className="h-4 w-4 flex items-center justify-center rounded-sm hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      <X size={10} />
                    </button>
                  </div>
                ) : (
                  <>
                    <button
                      type="button"
                      onClick={() => setSelected(v)}
                      className="flex items-center gap-2 px-3 py-1.5 rounded-sm text-meta text-left flex-1 min-w-0"
                      style={{ color: active ? 'var(--flow-text)' : 'var(--flow-text-muted)' }}
                    >
                      <Icon size={12} />
                      <span className="truncate">{v.name}</span>
                    </button>
                    <div className="flex items-center gap-0.5 pr-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onClick={(e) => toggleVisibility(v, e)}
                        disabled={updateView.isPending}
                        className="h-5 w-5 flex items-center justify-center rounded-sm hover:bg-[var(--flow-hover)] disabled:opacity-50"
                        style={{ color: 'var(--flow-text-dim)' }}
                        title={
                          isPrivate(v)
                            ? t('flow.views.makeSharedTooltip')
                            : t('flow.views.makePrivateTooltip')
                        }
                        aria-label={
                          isPrivate(v)
                            ? t('flow.views.makeSharedLabel')
                            : t('flow.views.makePrivateLabel')
                        }
                        aria-pressed={isPrivate(v)}
                      >
                        {isPrivate(v) ? <Lock size={9} /> : <Users size={9} />}
                      </button>
                      <button
                        type="button"
                        onClick={(e) => startEdit(v, e)}
                        className="h-5 w-5 flex items-center justify-center rounded-sm hover:bg-[var(--flow-hover)]"
                        style={{ color: 'var(--flow-text-dim)' }}
                        title={t('flow.views.renameTooltip')}
                        aria-label={t('flow.views.renameLabel')}
                      >
                        <Pencil size={9} />
                      </button>
                      <button
                        type="button"
                        onClick={(e) => handleDelete(v, e)}
                        className="h-5 w-5 flex items-center justify-center rounded-sm hover:bg-[var(--flow-hover)]"
                        style={{ color: 'var(--flow-text-dim)' }}
                        title={t('flow.views.deleteTooltip')}
                        aria-label={t('flow.views.deleteLabel')}
                        onMouseEnter={(e) =>
                          (e.currentTarget.style.color = 'var(--flow-danger-fg)')
                        }
                        onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--flow-text-dim)')}
                      >
                        <Trash2 size={9} />
                      </button>
                    </div>
                  </>
                )}
              </div>
            );
          })}

          {views.length === 0 && (
            <EmptyState
              icon={Eye}
              title={t('flow.views.emptyList')}
              action={{
                label: t('flow.views.addTooltip'),
                onClick: handleCreateView,
              }}
              size="compact"
            />
          )}
        </div>
      </aside>

      <div className="flex-1 overflow-y-auto">
        <div className="px-6 py-4" style={{ borderBottom: '1px solid var(--flow-border)' }}>
          <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {displayView?.name || t('flow.views.noViewSelected')}
          </h1>
          <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.views.taskCount', { count: tasks.length })}
          </p>
        </div>
        <motion.div
          key={displayView?.id}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.18 }}
        >
          {tasks.length === 0 && displayView && (
            <EmptyState
              icon={Eye}
              title={t('flow.views.emptyTasksForView')}
              action={{
                label: t('flow.shell.myTasks'),
                onClick: () => navigate({ to: '/my-issues' }),
              }}
            />
          )}
          {tasks.map((t) => (
            <TaskRow key={t.id} task={t} showProject />
          ))}
        </motion.div>
      </div>
      <ConfirmDialog
        open={pendingDelete !== null}
        onOpenChange={(o) => {
          if (!o) setPendingDelete(null);
        }}
        title={t('flow.views.deleteDialog.title')}
        description={
          <>
            {t('flow.views.deleteDialog.descriptionPrefix')} <strong>{pendingDelete?.name}</strong>{' '}
            {t('flow.views.deleteDialog.descriptionSuffix')}
          </>
        }
        confirmLabel={t('flow.views.deleteDialog.confirmLabel')}
        pending={deleteView.isPending}
        onConfirm={() => pendingDelete && performDelete(pendingDelete)}
      />
    </div>
  );
};

export default Views;
