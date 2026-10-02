import React from 'react';
import { useTranslation } from 'react-i18next';
import { Copy, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import type { ApiKey } from '../../hooks/queries/useApiKeys';

export interface ApiKeyRowProps {
  apiKey: ApiKey;
  onRevoke?: (id: string) => void;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export const ApiKeyRow: React.FC<ApiKeyRowProps> = ({ apiKey, onRevoke }) => {
  const { t } = useTranslation('common');
  const preview = `${apiKey.keyPrefix}••••••••`;

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(preview);
      toast.success(t('developer.apiKeys.previewCopied', 'Key preview copied'));
    } catch {
      toast.error(t('common.copyFailed', 'Failed to copy to clipboard'));
    }
  };

  return (
    <div className="flex items-center gap-4 px-4 py-3 rounded-lg transition-colors hover:bg-[var(--dash-hover)]">
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-semibold truncate" style={{ color: 'var(--dash-text)' }}>
            {apiKey.name}
          </p>
          <div className="flex items-center gap-1">
            {apiKey.scopes.map((s) => (
              <span
                key={s}
                className="px-1.5 py-0.5 rounded text-[9px] font-semibold uppercase"
                style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-muted)' }}
              >
                {s}
              </span>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-3 mt-1">
          <code className="text-xs font-mono" style={{ color: 'var(--dash-text-muted)' }}>
            {preview}
          </code>
          <button
            type="button"
            className="flex items-center gap-1 text-[11px] transition-opacity hover:opacity-70"
            style={{ color: 'var(--dash-text-dim)' }}
            onClick={() => void copy()}
          >
            <Copy size={10} /> Copy
          </button>
        </div>
      </div>
      <div className="hidden md:flex flex-col items-end shrink-0">
        <p className="text-[11px]" style={{ color: 'var(--dash-text-dim)' }}>
          {t('common.apiKey.created')}
        </p>
        <p className="text-xs font-medium" style={{ color: 'var(--dash-text-muted)' }}>
          {formatDate(apiKey.createdAt)}
        </p>
      </div>
      <button
        type="button"
        className="h-8 w-8 rounded-lg flex items-center justify-center transition-colors hover:bg-red-500/10"
        style={{ color: 'var(--status-cancelled)' }}
        onClick={() => onRevoke?.(apiKey.id)}
        aria-label={t('common.apiKey.revoke')}
      >
        <Trash2 size={14} />
      </button>
    </div>
  );
};

export default ApiKeyRow;
