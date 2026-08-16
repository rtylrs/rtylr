import React from 'react';
import { AlertCircle, AlertTriangle, CheckCircle2, Info } from 'lucide-react';
import AppBadge from './AppBadge';
import type { Notification } from '../../hooks/queries/useNotifications';

function formatRelative(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return 'now';
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h`;
  const days = Math.floor(hours / 24);
  return `${days}d`;
}

export interface NotificationItemProps {
  notification: Notification;
  onRead?: (id: string) => void;
  onClick?: (n: Notification) => void;
}

const TYPE_META = {
  info: { icon: Info, color: 'var(--accent)' },
  success: { icon: CheckCircle2, color: 'var(--status-active)' },
  warning: { icon: AlertTriangle, color: 'var(--status-trial)' },
  error: { icon: AlertCircle, color: 'var(--status-cancelled)' },
} as const;

export const NotificationItem: React.FC<NotificationItemProps> = ({
  notification,
  onRead,
  onClick,
}) => {
  const meta = TYPE_META[notification.type];
  const Icon = meta.icon;

  return (
    <div
      className="flex items-start gap-3 px-4 py-3 rounded-lg transition-colors hover:bg-[var(--dash-hover)] cursor-pointer"
      onClick={() => {
        if (!notification.read) onRead?.(notification.id);
        onClick?.(notification);
      }}
    >
      <div
        className="h-8 w-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5"
        style={{ background: 'var(--dash-hover)' }}
      >
        <Icon size={14} style={{ color: meta.color }} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <p className="text-sm font-semibold truncate" style={{ color: 'var(--dash-text)' }}>
            {notification.title}
          </p>
          {!notification.read && (
            <span
              className="h-1.5 w-1.5 rounded-full shrink-0 mt-1.5"
              style={{ background: 'var(--accent)' }}
            />
          )}
        </div>
        <p className="text-xs mt-0.5 line-clamp-2" style={{ color: 'var(--dash-text-muted)' }}>
          {notification.body}
        </p>
        <div className="flex items-center gap-2 mt-1.5">
          <AppBadge app={notification.app} size="xs" />
          <span className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
            {formatRelative(notification.createdAt)}
          </span>
        </div>
      </div>
    </div>
  );
};

export default NotificationItem;
