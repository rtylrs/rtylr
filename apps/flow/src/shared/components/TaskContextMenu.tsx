import React, { useEffect, useRef } from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { ExternalLink, Trash2, Copy } from 'lucide-react';
import { useDeleteTask } from '../../features/tasks/useTasks';
import { useCanDeleteResource } from '../lib/currentUser';
import { useTranslation } from 'react-i18next';

export interface ContextMenuState {
  x: number;
  y: number;
  taskId: string;
}

export const TaskContextMenu: React.FC<{
  menu: ContextMenuState;
  onClose: () => void;
}> = ({ menu, onClose }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const deleteTask = useDeleteTask();
  const canRemove = useCanDeleteResource();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const close = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    const closeKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('mousedown', close);
    document.addEventListener('keydown', closeKey);
    return () => {
      document.removeEventListener('mousedown', close);
      document.removeEventListener('keydown', closeKey);
    };
  }, [onClose]);

  const menuStyle: React.CSSProperties = {
    position: 'fixed',
    left: menu.x,
    top: menu.y,
    zIndex: 9999,
    minWidth: 160,
    background: 'var(--flow-card)',
    border: '1px solid var(--flow-border)',
    borderRadius: 8,
    boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
    padding: '4px 0',
  };

  const itemStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    width: '100%',
    padding: '6px 12px',
    fontSize: 12,
    cursor: 'pointer',
    background: 'transparent',
    border: 'none',
    color: 'var(--flow-text)',
    textAlign: 'left',
  };

  return createPortal(
    <div ref={ref} style={menuStyle} role="menu">
      <button
        style={itemStyle}
        onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--flow-hover)')}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
        onClick={() => {
          navigate({ to: `/task/${menu.taskId}` as string });
          onClose();
        }}
      >
        <ExternalLink size={12} style={{ color: 'var(--flow-text-muted)', flexShrink: 0 }} />
        Open task
      </button>
      <button
        style={itemStyle}
        onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--flow-hover)')}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
        onClick={() => {
          void navigator.clipboard.writeText(`${window.location.origin}/task/${menu.taskId}`);
          toast.success('Link copied');
          onClose();
        }}
      >
        <Copy size={12} style={{ color: 'var(--flow-text-muted)', flexShrink: 0 }} />
        Copy link
      </button>
      <div style={{ height: 1, background: 'var(--flow-border)', margin: '4px 0' }} />
      <button
        style={{
          ...itemStyle,
          color: canRemove ? '#ef4444' : 'var(--flow-text-dim)',
          cursor: canRemove ? 'pointer' : 'not-allowed',
        }}
        disabled={!canRemove}
        title={!canRemove ? t('common.deleteRestricted') : undefined}
        onMouseEnter={(e) => {
          if (canRemove) e.currentTarget.style.background = 'rgba(239,68,68,0.08)';
        }}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
        onClick={() => {
          if (!canRemove) return;
          deleteTask.mutate(menu.taskId, {
            onSuccess: () => toast.success('Task deleted'),
            onError: () => toast.error('Failed to delete task'),
          });
          onClose();
        }}
      >
        <Trash2
          size={12}
          style={{ color: canRemove ? '#ef4444' : 'var(--flow-text-dim)', flexShrink: 0 }}
        />
        Delete
      </button>
    </div>,
    document.body
  );
};
