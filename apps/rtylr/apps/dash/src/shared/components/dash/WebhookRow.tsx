import React from 'react';
import { useTranslation } from 'react-i18next';
import { Trash2, AlertTriangle, CheckCircle2, PauseCircle } from 'lucide-react';
import type { Webhook } from '../../hooks/queries/useWebhooks';

export interface WebhookRowProps {
  webhook: Webhook;
  onDelete?: (id: string) => void;
}

// Holds the key, not the label: this map is built at module load, before
// i18next has a language.
const STATUS_META = {
  active: { icon: CheckCircle2, color: 'var(--status-active)', labelKey: 'common.webhook.active' },
  paused: {
    icon: PauseCircle,
    color: 'var(--dash-text-muted)',
    labelKey: 'common.webhook.paused',
  },
  failing: {
    icon: AlertTriangle,
    color: 'var(--status-cancelled)',
    labelKey: 'common.webhook.failing',
  },
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
  const { t } = useTranslation();
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
            <Icon size={10} /> {t(meta.labelKey)}
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
        type="button"
        className="h-8 w-8 rounded-lg flex items-center justify-center transition-colors hover:bg-red-500/10 shrink-0"
        style={{ color: 'var(--status-cancelled)' }}
        onClick={() => onDelete?.(webhook.id)}
        aria-label={t('common.webhook.delete')}
      >
        <Trash2 size={14} />
      </button>
    </div>
  );
};

export default WebhookRow;
