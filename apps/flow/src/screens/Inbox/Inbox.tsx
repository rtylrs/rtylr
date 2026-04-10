import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  AtSign,
  CheckCircle2,
  Circle,
  MessageSquare,
  Bell,
  Filter,
  CheckCheck,
} from 'lucide-react';
import { cn, formatDate } from '../../lib/utils';
import { useUserStore } from '../../stores/userStore';
import { useFlowStore } from '../../stores/flowStore';

const ACCENT = '#64C5B9';

type FilterType = 'all' | 'unread' | 'assigned';

interface Notification {
  id: string;
  type: 'assigned' | 'mention' | 'comment' | 'status_change';
  read: boolean;
  title: string;
  description: string;
  project: string;
  projectColor: string;
  actor: string;
  actorInitials: string;
  timestamp: string;
  taskTitle: string;
}


const TYPE_ICONS: Record<Notification['type'], React.ReactNode> = {
  assigned:      <CheckCircle2 size={14} style={{ color: ACCENT }} />,
  mention:       <AtSign size={14} style={{ color: '#f59e0b' }} />,
  comment:       <MessageSquare size={14} style={{ color: '#3b82f6' }} />,
  status_change: <Circle size={14} style={{ color: '#22c55e' }} />,
};

function formatTimestamp(ts: string): string {
  const d = new Date(ts);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffH = Math.floor(diffMs / (1000 * 60 * 60));
  const diffD = Math.floor(diffH / 24);

  if (diffH < 1) return 'Just now';
  if (diffH < 24) return `${diffH}h ago`;
  if (diffD === 1) return 'Yesterday';
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

const NotificationRow: React.FC<{
  notification: Notification;
  onMarkRead: (id: string) => void;
}> = ({ notification, onMarkRead }) => {
  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, x: -12 }}
      transition={{ duration: 0.18 }}
      onClick={() => onMarkRead(notification.id)}
      className="flex items-start gap-3.5 px-5 py-4 cursor-pointer group hover:bg-[var(--flow-hover)] transition-colors relative"
      style={{ borderBottom: '1px solid var(--flow-border)' }}
    >
      {/* Unread indicator */}
      {!notification.read && (
        <span
          className="absolute left-2 top-1/2 -translate-y-1/2 h-1.5 w-1.5 rounded-full"
          style={{ background: ACCENT }}
        />
      )}

      {/* Actor avatar */}
      <span
        className="h-8 w-8 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0 mt-0.5"
        style={{
          background: `${notification.projectColor}22`,
          color: notification.projectColor,
        }}
      >
        {notification.actorInitials}
      </span>

      {/* Content */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-0.5">
          {TYPE_ICONS[notification.type]}
          <span
            className="text-[12px] font-semibold"
            style={{
              color: notification.read
                ? 'var(--flow-text-muted)'
                : 'var(--flow-text)',
            }}
          >
            {notification.title}
          </span>
          <span
            className="text-[11px] ml-auto shrink-0"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            {formatTimestamp(notification.timestamp)}
          </span>
        </div>

        <p
          className="text-[12px] mb-1"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {notification.description}
        </p>

        <div className="flex items-center gap-2">
          <span
            className="h-2 w-2 rounded-full"
            style={{ background: notification.projectColor }}
          />
          <span className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {notification.project}
          </span>
          <span className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            ·
          </span>
          <span
            className="text-[11px] truncate"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            {notification.taskTitle}
          </span>
        </div>
      </div>
    </motion.div>
  );
};

const Inbox: React.FC = () => {
  const { user } = useUserStore();
  const { notifications, markNotificationRead, markAllNotificationsRead } = useFlowStore();
  const [activeFilter, setActiveFilter] = useState<FilterType>('all');

  const markRead = (id: string) => {
    markNotificationRead(id);
  };

  const markAllRead = () => {
    markAllNotificationsRead();
  };

  const filtered = notifications.filter((n) => {
    if (activeFilter === 'unread') return !n.read;
    if (activeFilter === 'assigned') return n.type === 'assigned';
    return true;
  });

  const unreadCount = notifications.filter((n) => !n.read).length;

  return (
    <div className="flex flex-col h-full max-w-3xl mx-auto">
      {/* Header */}
      <div
        className="flex items-center justify-between px-6 py-5 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1
            className="text-[18px] font-bold"
            style={{ color: 'var(--flow-text)' }}
          >
            Inbox
          </h1>
          {unreadCount > 0 && (
            <p className="text-[12px] mt-0.5" style={{ color: 'var(--flow-text-muted)' }}>
              {unreadCount} unread notification{unreadCount !== 1 ? 's' : ''}
            </p>
          )}
        </div>

        <button
          onClick={markAllRead}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[12px] font-medium transition-colors hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <CheckCheck size={13} />
          Mark all read
        </button>
      </div>

      {/* Filters */}
      <div
        className="flex items-center gap-1 px-6 py-2 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        {(
          [
            { id: 'all', label: 'All' },
            { id: 'unread', label: 'Unread' },
            { id: 'assigned', label: 'Assigned to me' },
          ] as const
        ).map(({ id, label }) => (
          <button
            key={id}
            onClick={() => setActiveFilter(id)}
            className="px-3 py-1.5 rounded-md text-[12px] font-medium transition-colors"
            style={{
              background:
                activeFilter === id ? `${ACCENT}18` : 'transparent',
              color:
                activeFilter === id ? ACCENT : 'var(--flow-text-muted)',
            }}
          >
            {label}
            {id === 'unread' && unreadCount > 0 && (
              <span
                className="ml-1.5 px-1.5 py-0.5 rounded-full text-[10px] font-bold"
                style={{ background: ACCENT, color: '#0a0a0a' }}
              >
                {unreadCount}
              </span>
            )}
          </button>
        ))}
      </div>

      {/* Notifications list */}
      <div className="flex-1 overflow-y-auto">
        <AnimatePresence>
          {filtered.length === 0 ? (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="flex flex-col items-center justify-center py-20"
            >
              <Bell size={32} style={{ color: 'var(--flow-text-dim)' }} className="mb-3" />
              <p className="text-[14px] font-medium" style={{ color: 'var(--flow-text-muted)' }}>
                All caught up!
              </p>
              <p className="text-[12px] mt-1" style={{ color: 'var(--flow-text-dim)' }}>
                No notifications to show.
              </p>
            </motion.div>
          ) : (
            filtered.map((n) => (
              <NotificationRow key={n.id} notification={n} onMarkRead={markRead} />
            ))
          )}
        </AnimatePresence>
      </div>
    </div>
  );
};

export default Inbox;
