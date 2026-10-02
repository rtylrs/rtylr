import { EmptyState } from '../../../../shared/components/dash/EmptyState';
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

import type { SavedView } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import {
  useViews,
  useCreateView,
  useUpdateView,
  useDeleteView,
  useFilteredTasks,
} from './useViews';
import TaskRow from '../../shared/components/TaskRow';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import ViewFilterChips, { EMPTY_FILTERS, hasActiveFilters } from './ViewFilterChips';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

const ICON_MAP: Record<string, React.ComponentType<{ size?: number }>> = {
  flame: Flame,
  timer: Timer,
  'check-circle': CheckCircle2,
};

const Views: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const viewsQuery = useViews();
  const { data: views = [] } = viewsQuery;
  const createView = useCreateView();
  const updateView = useUpdateView();
  const deleteView = useDeleteView();
  const tasksQuery = useTasks({});
  const allTasks = (tasksQuery.data ?? []).map(apiTaskToTask).filter((task) => !task.archived);
  const [selected, setSelected] = useState<SavedView | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [pendingDelete, setPendingDelete] = useState<SavedView | null>(null);

  const displayView = views.find((view) => view.id === selected?.id) ?? views[0] ?? null;
  const tasks = useFilteredTasks(displayView, allTasks);
  const filtersActive = hasActiveFilters(displayView?.filters);

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
          toast.success(t('flowToasts.view.renamed'));
          setEditingId(null);
        },
        onError: () => toast.error(t('flowToasts.view.renameFailed')),
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
            next === 'private' ? t('flowToasts.view.madePrivate') : t('flowToasts.view.madeShared')
          ),
        onError: () => toast.error(t('flowToasts.view.visibilityFailed')),
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
      { onSuccess: () => toast.success(t('flowToasts.view.created')) }
    );
  };

  const performDelete = (v: SavedView) => {
    deleteView.mutate(v.id, {
      onSuccess: () => {
        toast.success(t('flowToasts.view.deleted'));
        if (selected?.id === v.id) setSelected(null);
        setPendingDelete(null);
      },
      onError: () => toast.error(t('flowToasts.view.deleteFailed')),
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
            className="text-flow-meta font-semibold uppercase tracking-wider"
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
                      className="flex-1 bg-transparent text-flow-meta"
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
                      className="flex items-center gap-2 px-3 py-1.5 rounded-sm text-flow-meta text-left flex-1 min-w-0"
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

          {/* A failed request emptied this list exactly as a first run does,
              so a 500 invited the user to create a view they already had. */}
          {hasQueryFailed(viewsQuery) ? (
            <ListLoadFailed onRetry={() => void viewsQuery.refetch()} />
          ) : null}

          {views.length === 0 && !viewsQuery.isLoading && !hasQueryFailed(viewsQuery) && (
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
          <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {displayView?.name || t('flow.views.noViewSelected')}
          </h1>
          {!tasksQuery.isLoading && !hasQueryFailed(tasksQuery) && (
            <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.views.taskCount', { count: tasks.length })}
            </p>
          )}
        </div>
        {displayView && (
          <ViewFilterChips
            filters={displayView.filters}
            disabled={updateView.isPending}
            onChange={(next) => updateView.mutate({ id: displayView.id, data: { filters: next } })}
          />
        )}
        <motion.div
          key={displayView?.id}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.18 }}
        >
          {tasksQuery.isLoading ? (
            <p className="px-6 py-10 text-flow-body" role="status">
              {t('flowCommon.loading')}
            </p>
          ) : hasQueryFailed(tasksQuery) ? (
            <ListLoadFailed onRetry={() => void tasksQuery.refetch()} />
          ) : (
            <>
              {tasks.length === 0 && displayView && filtersActive && (
                // G-E2: a filtered-empty result gets a plain sentence and a way to
                // clear the filters, no illustration. Sending the user to My tasks
                // was the only action available before, because there was no
                // filter control for the empty state to point at.
                <div className="px-6 py-10 text-center">
                  <p className="text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
                    {t('flow.views.emptyTasksForView')}
                  </p>
                  <button
                    type="button"
                    disabled={updateView.isPending}
                    onClick={() =>
                      updateView.mutate({ id: displayView.id, data: { filters: EMPTY_FILTERS } })
                    }
                    className="mt-3 inline-flex min-h-11 items-center rounded-md px-4 text-flow-body font-semibold disabled:opacity-50"
                    style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
                  >
                    {t('flow.views.filters.clearAll')}
                  </button>
                </div>
              )}
              {tasks.length === 0 && displayView && !filtersActive && (
                <EmptyState
                  icon={Eye}
                  title={t('flow.views.emptyTasksForView')}
                  action={{
                    label: t('flow.shell.myTasks'),
                    onClick: () => navigate({ to: '/flow/my-issues' }),
                  }}
                />
              )}
              {tasks.map((t) => (
                <TaskRow key={t.id} task={t} showProject />
              ))}
            </>
          )}
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
