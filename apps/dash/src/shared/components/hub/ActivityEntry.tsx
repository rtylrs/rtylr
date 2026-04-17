import React from 'react';
import { AlertCircle, AlertTriangle, Info } from 'lucide-react';
import AppBadge from './AppBadge';
import type { ActivityEntry as ActivityEntryType } from '../../hooks/queries/useActivity';

function formatRelative(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return 'just now';
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  const months = Math.floor(days / 30);
  return `${months}mo ago`;
}

export interface ActivityEntryProps {
  entry: ActivityEntryType;
  showIp?: boolean;
}

export const ActivityEntryRow: React.FC<ActivityEntryProps> = ({ entry, showIp }) => {
  const sev = entry.severity ?? 'info';
  const Icon = sev === 'critical' ? AlertCircle : sev === 'warning' ? AlertTriangle : Info;
  const iconColor = sev === 'critical' ? '#ef4444' : sev === 'warning' ? '#f59e0b' : '#4583C4';

  return (
    <div className="flex items-start gap-3 px-4 py-3 rounded-lg transition-colors hover:bg-[var(--hub-hover)]">
      <div
        className="h-8 w-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5"
        style={{ background: `${iconColor}15` }}
      >
        <Icon size={14} style={{ color: iconColor }} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-sm" style={{ color: 'var(--hub-text)' }}>
          <span className="font-semibold">{entry.actor}</span>
          <span className="font-normal" style={{ color: 'var(--hub-text-muted)' }}>
            {' '}
            {entry.action}{' '}
          </span>
          <span className="font-medium">{entry.target}</span>
        </p>
        <div className="flex items-center gap-2 mt-1">
          <AppBadge app={entry.app} size="xs" />
          <span className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
            {formatRelative(entry.createdAt)}
          </span>
          {showIp && entry.ip && (
            <span className="text-[11px] font-mono" style={{ color: 'var(--hub-text-dim)' }}>
              {entry.ip}
            </span>
          )}
        </div>
      </div>
    </div>
  );
};

export default ActivityEntryRow;
