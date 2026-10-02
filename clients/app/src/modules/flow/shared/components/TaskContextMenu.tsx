import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useOpenTask } from '../lib/useOpenTask';
import { createPortal } from 'react-dom';
import { useClickOutside } from '@voxire/utils';
import { toast } from 'sonner';
import { ExternalLink, Trash2, Copy } from 'lucide-react';
import { useDeleteTask } from '../../features/tasks/useTasks';
import { useCanDeleteResource } from '../lib/currentUser';
import { useTranslation } from 'react-i18next';
import ConfirmDialog from './ConfirmDialog';

export interface ContextMenuState {
  x: number;
  y: number;
  taskId: string;
  /** The task's project, so Delete can be gated on the user's role in it. */
  projectId: string;
  /** The task's title, so the delete confirmation can name what it removes. */
  taskTitle: string;
}

export const TaskContextMenu: React.FC<{
  menu: ContextMenuState;
  onClose: () => void;
}> = ({ menu, onClose }) => {
  const { t } = useTranslation();
  const openTask = useOpenTask();
  const deleteTask = useDeleteTask();
  const canRemove = useCanDeleteResource(menu.projectId);
  const ref = useRef<HTMLDivElement>(null);
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  // A menu's items are what the arrow keys move between, and role="menu" whose
  // children are plain buttons is invalid ARIA: the menu is announced but its
  // items are not exposed as items. Roving tabIndex plus arrow handling is what
  // makes it operable the way a menu is expected to be.
  const [activeIndex, setActiveIndex] = useState(0);
  const itemRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const itemCount = 3;

  const focusItem = useCallback((index: number): void => {
    const next = ((index % itemCount) + itemCount) % itemCount;
    setActiveIndex(next);
    itemRefs.current[next]?.focus();
  }, []);

  useEffect(() => {
    itemRefs.current[0]?.focus();
  }, []);

  const onMenuKeyDown = (event: React.KeyboardEvent<HTMLDivElement>): void => {
    if (event.key === 'ArrowDown') {
      event.preventDefault();
      focusItem(activeIndex + 1);
      return;
    }
    if (event.key === 'ArrowUp') {
      event.preventDefault();
      focusItem(activeIndex - 1);
      return;
    }
    if (event.key === 'Home') {
      event.preventDefault();
      focusItem(0);
      return;
    }
    if (event.key === 'End') {
      event.preventDefault();
      focusItem(itemCount - 1);
    }
  };

  useClickOutside(ref, onClose);
  useEffect(() => {
    const closeKey = (e: KeyboardEvent): void => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', closeKey);
    return () => document.removeEventListener('keydown', closeKey);
  }, [onClose]);

  const menuStyle: React.CSSProperties = {
    position: 'fixed',
    left: menu.x,
    top: menu.y,
    zIndex: 9999,
    minWidth: 160,
    background: 'var(--flow-card)',
    border: '1px solid var(--flow-border)',
    borderRadius: 'var(--radius-md)',
    // --flow-shadow is the popover and modal lift, and it is theme-aware. The
    // literal here was a heavy black drop that stayed heavy in light mode.
    boxShadow: 'var(--flow-shadow)',
    padding: '4px 0',
  };

  const itemStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    width: '100%',
    padding: '6px 12px',
    fontSize: 'var(--text-flow-meta)',
    cursor: 'pointer',
    background: 'transparent',
    border: 'none',
    color: 'var(--flow-text)',
    textAlign: 'left',
  };

  return createPortal(
    <>
      <div
        ref={ref}
        style={menuStyle}
        role="menu"
        aria-label={t('flow.taskMenu.label')}
        onKeyDown={onMenuKeyDown}
      >
        <button
          type="button"
          role="menuitem"
          ref={(node) => {
            itemRefs.current[0] = node;
          }}
          tabIndex={activeIndex === 0 ? 0 : -1}
          style={itemStyle}
          onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--flow-hover)')}
          onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
          onClick={() => {
            openTask(String(menu.taskId));
            onClose();
          }}
        >
          <ExternalLink size={12} style={{ color: 'var(--flow-text-muted)', flexShrink: 0 }} />
          {t('flow.taskMenu.openTask')}
        </button>
        <button
          type="button"
          role="menuitem"
          ref={(node) => {
            itemRefs.current[1] = node;
          }}
          tabIndex={activeIndex === 1 ? 0 : -1}
          style={itemStyle}
          onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--flow-hover)')}
          onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
          onClick={() => {
            void navigator.clipboard.writeText(`${window.location.origin}/task/${menu.taskId}`);
            toast.success(t('flowToasts.task.linkCopied'));
            onClose();
          }}
        >
          <Copy size={12} style={{ color: 'var(--flow-text-muted)', flexShrink: 0 }} />
          {t('flow.taskMenu.copyLink')}
        </button>
        <div style={{ height: 1, background: 'var(--flow-border)', margin: '4px 0' }} />
        <button
          type="button"
          role="menuitem"
          ref={(node) => {
            itemRefs.current[2] = node;
          }}
          tabIndex={activeIndex === 2 ? 0 : -1}
          style={{
            ...itemStyle,
            color: canRemove ? 'var(--flow-danger-fg)' : 'var(--flow-text-dim)',
            cursor: canRemove ? 'pointer' : 'not-allowed',
          }}
          // aria-disabled rather than disabled: a disabled button leaves the tab
          // order, which would put a hole in the menu's arrow navigation, and its
          // title tooltip stops showing. The click is guarded below instead.
          aria-disabled={!canRemove}
          title={!canRemove ? t('flowCommon.deleteRestricted') : undefined}
          onMouseEnter={(e) => {
            if (canRemove) e.currentTarget.style.background = 'rgba(239,68,68,0.08)';
          }}
          onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
          onClick={() => {
            if (!canRemove) return;
            // Deleting a task is immediate and has no undo, so it confirms
            // first, the way deleting a label or a board column already does.
            setConfirmingDelete(true);
          }}
        >
          <Trash2
            size={12}
            style={{
              color: canRemove ? 'var(--flow-danger-fg)' : 'var(--flow-text-dim)',
              flexShrink: 0,
            }}
          />
          {t('flowCommon.delete')}
        </button>
      </div>
      <ConfirmDialog
        open={confirmingDelete}
        onOpenChange={(open) => {
          setConfirmingDelete(open);
          if (!open) onClose();
        }}
        title={t('flow.taskMenu.deleteTitle')}
        description={
          <>
            {t('flow.taskMenu.deleteConfirmPrefix')} <strong>{menu.taskTitle}</strong>
            {t('flow.taskMenu.deleteConfirmSuffix')}
          </>
        }
        confirmLabel={t('flowCommon.delete')}
        cancelLabel={t('flowCommon.cancel')}
        pending={deleteTask.isPending}
        onConfirm={() => {
          deleteTask.mutate(menu.taskId, {
            onSuccess: () => toast.success(t('flowToasts.task.deleted')),
            onError: () => toast.error(t('flowToasts.task.deleteFailed')),
          });
          setConfirmingDelete(false);
          onClose();
        }}
      />
    </>,
    document.body
  );
};
