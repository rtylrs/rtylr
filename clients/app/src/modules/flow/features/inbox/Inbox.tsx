import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import {
  AtSign,
  CheckCircle2,
  Circle,
  MessageSquare,
  Bell,
  CheckCheck,
  Loader2,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

import { useTranslation } from '@voxire/i18n';
import type { TFunction } from 'i18next';
import {
  useNotifications,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
} from './useNotifications';
import type { ApiNotification } from '../../shared/services/flowApi';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { withAlpha, onTint } from '../../shared/lib/colorUtils';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import type { RowProps } from '../../../../shared/hooks/useVirtualRows';
import { VirtualList } from '../../../../shared/components/VirtualList';

type FilterType = 'all' | 'unread' | 'assigned';

const TYPE_ICONS: Record<ApiNotification['type'], React.ReactNode> = {
  assigned: <CheckCircle2 size={14} style={{ color: 'var(--accent)' }} />,
  mention: <AtSign size={14} style={{ color: 'var(--flow-warn-fg)' }} />,
  comment: <MessageSquare size={14} style={{ color: 'var(--flow-info-fg)' }} />,
  status_change: <Circle size={14} style={{ color: 'var(--flow-status-approved)' }} />,
  due_soon: <Bell size={14} style={{ color: 'var(--flow-danger-fg)' }} />,
};

function formatTimestamp(ts: string, t: TFunction, lang: string): string {
  const d = new Date(ts);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffH = Math.floor(diffMs / (1000 * 60 * 60));
  const diffD = Math.floor(diffH / 24);

  if (diffH < 1) return t('flowInbox.time.justNow');
  if (diffH < 24) return t('flowInbox.time.hoursAgo', { count: diffH });
  if (diffD === 1) return t('flowInbox.time.yesterday');
  return d.toLocaleDateString(lang || 'en-US', { month: 'short', day: 'numeric' });
}

const NotificationRow: React.FC<{
  notification: ApiNotification;
  onOpen: (notification: ApiNotification) => void;
  /** From a windowed list (see VirtualList). */
  rowProps?: RowProps;
}> = ({ notification, onOpen, rowProps }) => {
  const { t, i18n } = useTranslation();
  // Windowed rows mount as they scroll into view; no entrance or layout animation.
  const windowed = rowProps?.['data-index'] !== undefined;
  return (
    <motion.button
      {...rowProps}
      type="button"
      layout={!windowed}
      initial={windowed ? false : { opacity: 0, y: 6 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, x: -12 }}
      transition={{ duration: 0.18 }}
      onClick={() => onOpen(notification)}
      className="flex w-full items-start gap-3.5 px-5 py-4 text-left cursor-pointer group hover:bg-[var(--flow-hover)] transition-colors relative"
      style={{ borderBottom: '1px solid var(--flow-border)', background: 'transparent' }}
    >
      {!notification.read && (
        <span
          className="absolute left-2 top-1/2 -translate-y-1/2 h-1.5 w-1.5 rounded-full"
          style={{ background: 'var(--flow-action-bg)' }}
        />
      )}

      <span
        className="h-8 w-8 rounded-full flex items-center justify-center text-flow-meta font-bold shrink-0 mt-0.5"
        style={{
          // See Calendar: the tone is set once and read back, which also makes it
          // observable in tests.
          ...(notification.projectColor ? { ['--tone' as string]: notification.projectColor } : {}),
          background: notification.projectColor
            ? withAlpha('var(--tone)', 13)
            : 'var(--accent-muted)',
          color: notification.projectColor ? onTint('var(--tone)') : 'var(--accent-on-muted)',
        }}
      >
        {notification.actorInitials || '?'}
      </span>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-0.5">
          {TYPE_ICONS[notification.type]}
          <span
            className="text-flow-meta font-semibold"
            style={{
              color: notification.read ? 'var(--flow-text-muted)' : 'var(--flow-text)',
            }}
          >
            {notification.title}
          </span>
          <span
            className="text-flow-meta ml-auto shrink-0"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            {formatTimestamp(notification.timestamp, t, i18n.language)}
          </span>
        </div>

        <p className="text-flow-meta mb-1" style={{ color: 'var(--flow-text-muted)' }}>
          {notification.description}
        </p>

        <div className="flex items-center gap-2">
          {notification.projectColor && (
            <span
              className="h-2 w-2 rounded-full"
              style={{ background: notification.projectColor }}
            />
          )}
          <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {notification.project}
          </span>
          {notification.taskTitle && (
            <>
              <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                ·
              </span>
              <span className="text-flow-meta truncate" style={{ color: 'var(--flow-text-dim)' }}>
                {notification.taskTitle}
              </span>
            </>
          )}
        </div>
      </div>
    </motion.button>
  );
};

const Inbox: React.FC = () => {
  const { t } = useTranslation();
  const notificationsQuery = useNotifications();
  const { data: notifications = [], isLoading, isPending, refetch } = notificationsQuery;
  const initialLoading = isPending || isLoading;
  const queryFailed = hasQueryFailed(notificationsQuery);
  const loadFailed = queryFailed && notifications.length === 0;
  const markRead = useMarkNotificationRead();
  const markAll = useMarkAllNotificationsRead();
  const openTask = useOpenTask();
  const [activeFilter, setActiveFilter] = useState<FilterType>('all');

  const filtered = notifications.filter((n) => {
    if (activeFilter === 'unread') return !n.read;
    if (activeFilter === 'assigned') return n.type === 'assigned';
    return true;
  });

  const unreadCount = notifications.filter((n) => !n.read).length;

  const filterDefs = [
    { id: 'all' as const, label: t('flowInbox.filters.all') },
    { id: 'unread' as const, label: t('flowInbox.filters.unread') },
    { id: 'assigned' as const, label: t('flowInbox.filters.assigned') },
  ];

  return (
    <div className="flex flex-col h-full max-w-3xl mx-auto">
      <HelmetContainer title={`${t('flowInbox.title')} · Flow`} />
      <div
        className="flex items-center justify-between px-6 py-5 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-flow-title font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flowInbox.title')}
          </h1>
          {unreadCount > 0 && (
            <p className="text-flow-meta mt-0.5" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flowInbox.unreadCount', { count: unreadCount })}
            </p>
          )}
        </div>

        <button
          type="button"
          onClick={() => markAll.mutate()}
          disabled={markAll.isPending || unreadCount === 0}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-flow-meta font-medium transition-colors hover:bg-[var(--flow-hover)] disabled:opacity-40"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {markAll.isPending ? (
            <Loader2 size={13} className="animate-spin" />
          ) : (
            <CheckCheck size={13} />
          )}
          {t('flowInbox.markAllRead')}
        </button>
      </div>

      <div
        className="flex items-center gap-1 px-6 py-2 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        {filterDefs.map(({ id, label }) => (
          <button
            type="button"
            key={id}
            onClick={() => setActiveFilter(id)}
            className="px-3 py-1.5 rounded-md text-flow-meta font-medium transition-colors"
            style={{
              background: activeFilter === id ? 'var(--accent-muted)' : 'transparent',
              color: activeFilter === id ? 'var(--accent-on-muted)' : 'var(--flow-text-muted)',
            }}
          >
            {label}
            {id === 'unread' && unreadCount > 0 && (
              <span
                className="ml-1.5 px-1.5 py-0.5 rounded-full text-flow-meta font-bold"
                style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
              >
                {unreadCount}
              </span>
            )}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto">
        {loadFailed ? (
          <div className="py-10">
            <ListLoadFailed onRetry={() => void refetch()} />
          </div>
        ) : initialLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 size={20} className="animate-spin" style={{ color: 'var(--flow-text-dim)' }} />
          </div>
        ) : (
          <>
            {queryFailed && notifications.length > 0 && (
              <div
                className="mx-5 my-3 flex items-center justify-between gap-4 rounded-lg border px-4 py-3 text-flow-meta"
                role="status"
                style={{ borderColor: 'var(--flow-border)', color: 'var(--flow-text-muted)' }}
              >
                <span>{t('flow.loadFailed.description')}</span>
                <button
                  type="button"
                  className="shrink-0 font-semibold hover:underline"
                  onClick={() => void refetch()}
                >
                  {t('flow.loadFailed.retry')}
                </button>
              </div>
            )}
            <AnimatePresence>
              {filtered.length === 0 ? (
                <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="py-10">
                  <EmptyState
                    icon={Bell}
                    title={t('flowInbox.emptyTitle')}
                    description={t('flowInbox.emptyDescription')}
                    action={
                      activeFilter !== 'all'
                        ? {
                            label: t('flowInbox.filters.all'),
                            onClick: () => setActiveFilter('all'),
                          }
                        : {
                            label: t('flowInbox.refreshCta'),
                            onClick: () => void refetch(),
                          }
                    }
                  />
                </motion.div>
              ) : (
                <VirtualList
                  key="list"
                  count={filtered.length}
                  estimateSize={73}
                  renderRow={(index, rowProps) => (
                    <NotificationRow
                      key={filtered[index].id}
                      rowProps={rowProps}
                      notification={filtered[index]}
                      onOpen={(notification) => {
                        markRead.mutate(notification.id);
                        if (notification.taskId) openTask(notification.taskId);
                      }}
                    />
                  )}
                />
              )}
            </AnimatePresence>
          </>
        )}
      </div>
    </div>
  );
};

export default Inbox;
