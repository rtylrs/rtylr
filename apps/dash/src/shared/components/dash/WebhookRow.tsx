import React from 'react';
import { Trash2, AlertTriangle, CheckCircle2, PauseCircle } from 'lucide-react';
import type { Webhook } from '../../hooks/queries/useWebhooks';

export interface WebhookRowProps {
  webhook: Webhook;
  onDelete?: (id: string) => void;
}

const STATUS_META = {
  active: { icon: CheckCircle2, color: '#10b981', label: 'Active' },
  paused: { icon: PauseCircle, color: '#6b7280', label: 'Paused' },
  failing: { icon: AlertTriangle, color: '#ef4444', label: 'Failing' },
} as const;

function formatRelative(iso?: string): string {
  if (!iso) return 'never';
  const diff = Date.now() - new Date(iso).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export const WebhookRow: React.FC<WebhookRowProps> = ({ webhook, onDelete }) => {
  const meta = STATUS_META[webhook.status];
  const Icon = meta.icon;

  return (
    <div className="flex items-start gap-4 px-4 py-3 rounded-lg transition-colors hover:bg-[var(--dash-hover)]">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <code
            className="text-sm font-mono font-semibold truncate"
            style={{ color: 'var(--dash-text)' }}
          >
            {webhook.url}
          </code>
          <span
            className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase shrink-0"
            style={{ background: `${meta.color}15`, color: meta.color }}
          >
            <Icon size={10} /> {meta.label}
          </span>
        </div>
        <div className="flex items-center gap-2 mt-1.5 flex-wrap">
          {webhook.events.map((e) => (
            <span
              key={e}
              className="px-1.5 py-0.5 rounded text-[10px] font-mono"
              style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
            >
              {e}
            </span>
          ))}
        </div>
        <p className="text-[11px] mt-1.5" style={{ color: 'var(--dash-text-dim)' }}>
          Last delivery {formatRelative(webhook.lastDelivery)} · {webhook.secretPreview}
        </p>
      </div>
      <button
        className="h-8 w-8 rounded-lg flex items-center justify-center transition-colors hover:bg-red-500/10 shrink-0"
        style={{ color: '#ef4444' }}
        onClick={() => onDelete?.(webhook.id)}
        aria-label="Delete webhook"
      >
        <Trash2 size={14} />
      </button>
    </div>
  );
};

export default WebhookRow;
