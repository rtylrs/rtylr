import { PageHeader } from '../../shared/components/dash/PageHeader';
import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { resolveAppUrl } from '@voxire/config';
import { useTranslation } from 'react-i18next';
import { Bell, CheckCheck } from 'lucide-react';
import { EmptyState } from '../../shared/components/dash/EmptyState';
import { Button } from '../../shared/components/ui';
import { NotificationItem, SkeletonList } from '../../shared/components/dash';
import {
  useNotificationPages,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
} from '../../shared/hooks/queries/useNotifications';

const TABS = ['all', 'unread'] as const;

type NotificationTarget = {
  path: string;
  isCurrentOrigin: boolean;
  href: string;
};

export function resolveNotificationTarget(
  link: string | undefined,
  originOverride?: string,
  dashOriginOverride?: string
): NotificationTarget | null {
  if (!link) return null;

  try {
    const currentOrigin = originOverride ?? window.location.origin;
    const dashOrigin = dashOriginOverride ?? new URL(resolveAppUrl('dash')).origin;
    const parsed = new URL(link, currentOrigin);
    if (
      (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') ||
      parsed.username ||
      parsed.password ||
      (parsed.origin !== currentOrigin && parsed.origin !== dashOrigin)
    ) {
      return null;
    }

    return {
      path: `${parsed.pathname}${parsed.search}${parsed.hash}`,
      isCurrentOrigin: parsed.origin === currentOrigin,
      href: parsed.toString(),
    };
  } catch {
    return null;
  }
}

const Notifications: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
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

  const handleNotificationClick = (notification: (typeof data)[number]) => {
    const target = resolveNotificationTarget(notification.link ?? notification.href);
    if (!target) return;
    if (target.isCurrentOrigin) {
      void navigate({ to: target.path });
    } else {
      window.location.assign(target.href);
    }
  };

  return (
    <>
      <HelmetContainer title="Notifications · rtylr" />
      <div className="portal-management-page space-y-5 max-w-3xl">
        <PageHeader
          title={t('common.nav.notifications')}
          description={t('common.management.notificationCount', {
            unread: unreadCount,
            total: data.length,
          })}
          actions={
            unreadCount > 0 && (
              <Button variant="outline" onClick={() => markAll.mutate()}>
                <CheckCheck />
                {t('common.management.markAllRead')}
              </Button>
            )
          }
        />

        <div className="portal-filter-list" aria-label={t('common.management.notificationFilter')}>
          {TABS.map((tabId) => (
            <button
              type="button"
              key={tabId}
              onClick={() => setTab(tabId)}
              className="portal-filter-pill"
              aria-pressed={tab === tabId}
            >
              {t(
                tabId === 'all'
                  ? 'common.management.allNotifications'
                  : 'common.management.unreadNotifications'
              )}
              {tabId === 'unread' && unreadCount > 0 && (
                <span className="ms-1.5">({unreadCount})</span>
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
          ) : (
            <>
              {filtered.length === 0 ? (
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
                      onClick={handleNotificationClick}
                    />
                  ))}
                </div>
              )}
              {hasNextPage && (
                <div className="px-2 pb-2 pt-1 flex justify-center">
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
            </>
          )}
        </div>
      </div>
    </>
  );
};

export default Notifications;
