import React from 'react';
import { useTranslation } from 'react-i18next';
import type { RealtimeStatus } from '@voxire/api';
import { cn } from '../lib/utils';

interface RealtimeConnectionIndicatorProps {
  status: RealtimeStatus;
  /** Dot-only rendering for tight header space (mobile top bar). */
  compact?: boolean;
  className?: string;
}

/** Show interrupted updates while allowing ordinary API requests to continue. */
const RealtimeConnectionIndicator = ({
  status,
  compact,
  className,
}: RealtimeConnectionIndicatorProps): React.JSX.Element | null => {
  const { t } = useTranslation();

  if (status !== 'reconnecting' && status !== 'disconnected') return null;

  const label =
    status === 'reconnecting'
      ? t('flow.shell.realtimeReconnecting')
      : t('flow.shell.realtimeDisconnected');

  const dot = (
    <span
      className="h-1.5 w-1.5 shrink-0 rounded-full"
      style={{ background: 'var(--flow-status-pending)' }}
      aria-hidden="true"
    />
  );

  if (compact) {
    return (
      <div
        role="status"
        aria-live="polite"
        aria-label={label}
        title={label}
        className={cn('flex h-10 w-10 shrink-0 items-center justify-center rounded-lg', className)}
      >
        {dot}
      </div>
    );
  }

  return (
    <div
      role="status"
      aria-live="polite"
      title={label}
      className={cn(
        'flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-flow-meta font-medium',
        className
      )}
      style={{
        background: 'var(--flow-status-pending-bg)',
        border: '1px solid var(--flow-status-pending-border)',
        color: 'var(--flow-status-pending)',
      }}
    >
      {dot}
      <span className="truncate">{label}</span>
    </div>
  );
};

export default RealtimeConnectionIndicator;
