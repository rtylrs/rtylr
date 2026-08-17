import React, { useState } from 'react';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { Eye, Plus, Flame, Timer, CheckCircle2, Pencil, Trash2, Check, X } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
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

  const handleDelete = (v: SavedView, e: React.MouseEvent) => {
    e.stopPropagation();
    setPendingDelete(v);
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
      <HelmetContainer title="Views · rtylr" />
      <aside
        className="w-64 shrink-0 px-3 py-4"
        style={{
          background: 'var(--flow-surface)',
          borderRight: '1px solid var(--flow-border)',
        }}
      >
        <div className="flex items-center justify-between px-2 mb-3">
          <h2
            className="text-[11px] font-semibold uppercase tracking-wider"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            Saved Views
          </h2>
          <button
            onClick={() =>
              createView.mutate(
                { name: 'New view', icon: 'eye', filters: {} },
                { onSuccess: () => toast.success(t('toasts.view.created')) }
              )
            }
            disabled={createView.isPending}
            className="h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
            title="New view"
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
                className="group flex items-center gap-1 rounded"
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
                      className="flex-1 bg-transparent text-[12px] outline-none"
                      style={{ color: 'var(--flow-text)' }}
                    />
                    <button
                      onClick={() => commitEdit(v)}
                      className="h-4 w-4 flex items-center justify-center rounded hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--accent)' }}
                    >
                      <Check size={10} />
                    </button>
                    <button
                      onClick={() => setEditingId(null)}
                      className="h-4 w-4 flex items-center justify-center rounded hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      <X size={10} />
                    </button>
                  </div>
                ) : (
                  <>
                    <button
                      onClick={() => setSelected(v)}
                      className="flex items-center gap-2 px-3 py-1.5 rounded text-[12px] text-left flex-1 min-w-0"
                      style={{ color: active ? 'var(--flow-text)' : 'var(--flow-text-muted)' }}
                    >
                      <Icon size={12} />
                      <span className="truncate">{v.name}</span>
                    </button>
                    <div className="flex items-center gap-0.5 pr-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => startEdit(v, e)}
                        className="h-5 w-5 flex items-center justify-center rounded hover:bg-[var(--flow-hover)]"
                        style={{ color: 'var(--flow-text-dim)' }}
                        title="Rename"
                      >
                        <Pencil size={9} />
                      </button>
                      <button
                        onClick={(e) => handleDelete(v, e)}
                        className="h-5 w-5 flex items-center justify-center rounded hover:bg-[var(--flow-hover)]"
                        style={{ color: 'var(--flow-text-dim)' }}
                        title="Delete"
                        onMouseEnter={(e) => (e.currentTarget.style.color = '#ef4444')}
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
            <p className="text-[11px] px-2 py-2" style={{ color: 'var(--flow-text-dim)' }}>
              No saved views yet. Click + to create one.
            </p>
          )}
        </div>
      </aside>

      <div className="flex-1 overflow-y-auto">
        <div className="px-6 py-4" style={{ borderBottom: '1px solid var(--flow-border)' }}>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {displayView?.name || 'No view selected'}
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {tasks.length} task{tasks.length !== 1 ? 's' : ''}
          </p>
        </div>
        <motion.div
          key={displayView?.id}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.18 }}
        >
          {tasks.length === 0 && displayView && (
            <div className="flex items-center justify-center py-20">
              <p className="text-[13px]" style={{ color: 'var(--flow-text-dim)' }}>
                No tasks match this view's filters.
              </p>
            </div>
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
        title="Delete saved view"
        description={
          <>
            Delete the <strong>{pendingDelete?.name}</strong> view? The filters are lost; tasks
            themselves aren't affected.
          </>
        }
        confirmLabel="Delete view"
        pending={deleteView.isPending}
        onConfirm={() => pendingDelete && performDelete(pendingDelete)}
      />
    </div>
  );
};

export default Views;
