import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { Bell, CheckCheck } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { Button } from '../../shared/components/ui';
import { Breadcrumbs, NotificationItem, SkeletonList } from '../../shared/components/dash';
import {
  useNotificationPages,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
} from '../../shared/hooks/queries/useNotifications';

const TABS = ['all', 'unread'] as const;

const Notifications: React.FC = () => {
  const { t } = useTranslation();
  const [tab, setTab] = useState<(typeof TABS)[number]>('all');
  const {
    data: pages,
    isLoading,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    refetch,
  } = useNotificationPages();
  const markRead = useMarkNotificationRead();
  const markAll = useMarkAllNotificationsRead();

  const data = pages?.pages.flatMap((page) => page.notifications) ?? [];
  const filtered = tab === 'unread' ? data.filter((n) => !n.read) : data;
  const unreadCount = data.filter((n) => !n.read).length;

  return (
    <>
      <HelmetContainer title="Notifications · rtylr" />
      <div className="space-y-5 max-w-3xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div
              className="h-9 w-9 rounded-xl flex items-center justify-center"
              style={{ background: 'rgba(69,131,196,0.15)' }}
            >
              <Bell size={16} style={{ color: '#4583C4' }} />
            </div>
            <div>
              <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
                Notifications
              </h1>
              <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
                {unreadCount} unread · {data.length} total
              </p>
            </div>
          </div>
          {unreadCount > 0 && (
            <Button variant="secondary" size="sm" onClick={() => markAll.mutate()}>
              <CheckCheck size={14} />
              Mark all read
            </Button>
          )}
        </div>

        <div className="flex items-center gap-1">
          {TABS.map((t) => (
            <button
              type="button"
              key={t}
              onClick={() => setTab(t)}
              className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
              style={{
                background: tab === t ? 'rgba(69,131,196,0.15)' : 'var(--dash-hover)',
                color: tab === t ? '#4583C4' : 'var(--dash-text-muted)',
                border:
                  tab === t ? '1px solid rgba(69,131,196,0.3)' : '1px solid var(--dash-border)',
              }}
            >
              {t}
              {t === 'unread' && unreadCount > 0 && (
                <span className="ml-1.5 text-[10px]">({unreadCount})</span>
              )}
            </button>
          ))}
        </div>

        <div
          className="rounded-xl overflow-hidden"
          style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
        >
          {isLoading ? (
            <div className="p-2">
              <SkeletonList rows={5} />
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-4">
              <EmptyState
                icon={Bell}
                title={t('notifications.emptyAllCaughtUpTitle')}
                description={
                  tab === 'unread'
                    ? t('notifications.emptyNoUnreadDescription')
                    : t('notifications.emptyNoneDescription')
                }
                action={
                  tab === 'unread'
                    ? { label: t('notifications.viewAll'), onClick: () => setTab('all') }
                    : { label: t('common.refresh'), onClick: () => void refetch() }
                }
              />
            </div>
          ) : (
            <div className="p-2">
              {filtered.map((n) => (
                <NotificationItem
                  key={n.id}
                  notification={n}
                  onRead={(id) => markRead.mutate(id)}
                />
              ))}
              {hasNextPage && (
                <div className="pt-2 flex justify-center">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => void fetchNextPage()}
                    disabled={isFetchingNextPage}
                  >
                    {isFetchingNextPage ? 'Loading…' : 'Load more'}
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </>
  );
};

export default Notifications;
