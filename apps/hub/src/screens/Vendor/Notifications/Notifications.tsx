import React, { useState } from 'react';
import { Bell, CheckCheck } from 'lucide-react';
import { Button } from '../../../components/ui';
import { Breadcrumbs, EmptyState, NotificationItem, SkeletonList } from '../../../components/hub';
import {
  useNotifications,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
} from '../../../hooks/queries/useNotifications';

const TABS = ['all', 'unread'] as const;

const Notifications: React.FC = () => {
  const [tab, setTab] = useState<(typeof TABS)[number]>('all');
  const { data = [], isLoading } = useNotifications();
  const markRead = useMarkNotificationRead();
  const markAll = useMarkAllNotificationsRead();

  const filtered = tab === 'unread' ? data.filter((n) => !n.read) : data;
  const unreadCount = data.filter((n) => !n.read).length;

  return (
    <div className="space-y-5 max-w-3xl">
      
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="h-9 w-9 rounded-xl flex items-center justify-center" style={{ background: 'rgba(69,131,196,0.15)' }}>
            <Bell size={16} style={{ color: '#4583C4' }} />
          </div>
          <div>
            <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Notifications</h1>
            <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
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
            key={t}
            onClick={() => setTab(t)}
            className="px-2.5 py-1 rounded-md text-[11px] font-semibold uppercase tracking-wider transition-colors"
            style={{
              background: tab === t ? 'rgba(69,131,196,0.15)' : 'var(--hub-hover)',
              color:      tab === t ? '#4583C4' : 'var(--hub-text-muted)',
              border:     tab === t ? '1px solid rgba(69,131,196,0.3)' : '1px solid var(--hub-border)',
            }}
          >
            {t}
            {t === 'unread' && unreadCount > 0 && <span className="ml-1.5 text-[10px]">({unreadCount})</span>}
          </button>
        ))}
      </div>

      <div className="rounded-xl overflow-hidden" style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}>
        {isLoading ? (
          <div className="p-2"><SkeletonList rows={5} /></div>
        ) : filtered.length === 0 ? (
          <div className="p-4">
            <EmptyState icon={Bell} title="All caught up" description={tab === 'unread' ? 'No unread notifications.' : 'No notifications yet.'} />
          </div>
        ) : (
          <div className="p-2">
            {filtered.map((n) => <NotificationItem key={n.id} notification={n} onRead={(id) => markRead.mutate(id)} />)}
          </div>
        )}
      </div>
    </div>
  );
};

export default Notifications;
