import React from 'react';
import { motion } from 'motion/react';
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
  index?: number;
}

export const ActivityEntryRow: React.FC<ActivityEntryProps> = ({ entry, showIp, index = 0 }) => {
  const sev = entry.severity ?? 'info';
  const Icon = sev === 'critical' ? AlertCircle : sev === 'warning' ? AlertTriangle : Info;
  const iconColor =
    sev === 'critical'
      ? 'var(--status-cancelled)'
      : sev === 'warning'
        ? 'var(--status-trial)'
        : 'var(--accent)';

  return (
    <motion.div
      className="flex items-start gap-3 px-4 py-3 rounded-lg hover:bg-[var(--dash-hover)]"
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.26, ease: [0.22, 1, 0.36, 1], delay: Math.min(index, 12) * 0.03 }}
    >
      <div
        className="h-8 w-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5"
        style={{ background: 'var(--dash-hover)' }}
      >
        <Icon size={14} style={{ color: iconColor }} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-sm" style={{ color: 'var(--dash-text)' }}>
          <span className="font-semibold">{entry.actor}</span>
          <span className="font-normal" style={{ color: 'var(--dash-text-muted)' }}>
            {' '}
            {entry.action}{' '}
          </span>
          <span className="font-medium">{entry.target}</span>
        </p>
        <div className="flex items-center gap-2 mt-1">
          <AppBadge app={entry.app} size="xs" />
          <span className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
            {formatRelative(entry.createdAt)}
          </span>
          {showIp && entry.ip && (
            <span className="text-[11px] font-mono" style={{ color: 'var(--dash-text-dim)' }}>
              {entry.ip}
            </span>
          )}
        </div>
      </div>
    </motion.div>
  );
};

export default ActivityEntryRow;
