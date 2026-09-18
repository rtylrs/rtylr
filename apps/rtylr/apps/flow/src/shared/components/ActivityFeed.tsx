import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  Plus,
  ArrowRight,
  User,
  Flag,
  Calendar,
  MessageCircle,
  Tag,
  Paperclip,
} from 'lucide-react';
import { ActivityEntry } from '../stores/flowStore';

interface ActivityFeedProps {
  entries: ActivityEntry[];
}

const ICONS: Record<ActivityEntry['type'], React.ComponentType<{ size?: number }>> = {
  created: Plus,
  status_change: ArrowRight,
  assignee_change: User,
  priority_change: Flag,
  due_date_change: Calendar,
  comment: MessageCircle,
  label_added: Tag,
  attachment: Paperclip,
};

function timeAgo(ts: string): string {
  const d = new Date(ts);
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return Math.floor(hours / 24) + 'd ago';
}

function useActivityLabel(): (a: ActivityEntry) => string {
  const { t } = useTranslation();
  return (a: ActivityEntry): string => {
    switch (a.type) {
      case 'created':
        return t('common.activity.created');
      case 'status_change':
        return t('common.activity.statusChanged', { from: a.from, to: a.to });
      case 'assignee_change':
        return t('common.activity.assigneeChanged', { to: a.to });
      case 'priority_change':
        return t('common.activity.priorityChanged', { from: a.from, to: a.to });
      case 'due_date_change':
        return t('common.activity.dueDateChanged', { to: a.to });
      case 'comment':
        return t('common.activity.commented');
      case 'label_added':
        return t('common.activity.labelAdded', { detail: a.detail });
      case 'attachment':
        return t('common.activity.attachment', { detail: a.detail });
    }
  };
}

export const ActivityFeed: React.FC<ActivityFeedProps> = ({ entries }) => {
  const { t } = useTranslation();
  const getActivityLabel = useActivityLabel();

  if (entries.length === 0) {
    return (
      <div className="text-meta text-center py-4" style={{ color: 'var(--flow-text-dim)' }}>
        {t('common.activity.empty', 'No activity yet.')}
      </div>
    );
  }
  return (
    <div className="space-y-3">
      {entries.map((a) => {
        const Icon = ICONS[a.type];
        return (
          <div key={a.id} className="flex items-start gap-2.5">
            <span
              className="h-5 w-5 rounded-full flex items-center justify-center shrink-0 mt-0.5"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
            >
              <Icon size={10} />
            </span>
            <div className="flex-1 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              <span className="font-semibold" style={{ color: 'var(--flow-text)' }}>
                {a.actor}
              </span>{' '}
              {getActivityLabel(a)}
              <span className="ml-2" style={{ color: 'var(--flow-text-dim)' }}>
                {timeAgo(a.createdAt)}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default ActivityFeed;
