import React from 'react';
import { useTranslation } from 'react-i18next';
import { Users, AlertCircle } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { Avatar, EmptyState } from '@voxire/ui';
import { CardSkeleton } from '../../shared/components/Skeleton';
import type { PresenceEntry } from '../../shared/schemas/presence';
import { usePresence } from './usePresence';

function initials(name: string): string {
  return name
    .split(' ')
    .filter(Boolean)
    .map((part) => part[0])
    .slice(0, 2)
    .join('')
    .toUpperCase();
}

const statusTone: Record<string, string> = {
  on_leave: 'var(--flow-text-muted)',
  not_started: 'var(--flow-text-muted)',
};

/**
 * Business-wide presence, backed by /flow/presence/today - gated
 * server-side by view_directory (onion PR #267). Mode/status only, no
 * clocked/logged/gap minutes: a company-wide view watched by every
 * employee is a different exposure than the workspace-admin-gated Team view.
 */
const PresenceScreen = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading, isError, refetch } = usePresence();

  const statusLabel = (status: string): string => {
    if (status === 'on_leave') return t('flow.presence.onLeave');
    if (status === 'not_started') return t('flow.presence.notStarted');
    return t(`flow.time.mode.${status}`);
  };

  return (
    <>
      <HelmetContainer title={`${t('flow.presence.heading')} · Flow`} />
      <div className="mx-auto max-w-2xl p-4 sm:p-6">
        <h1 className="mb-4 text-lg font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.presence.heading')}
        </h1>

        {isLoading && <CardSkeleton />}

        {!isLoading && isError && (
          <EmptyState
            icon={AlertCircle}
            title={t('flow.presence.loadFailedTitle')}
            description={t('flow.presence.loadFailedDescription')}
            action={{
              label: t('flow.presence.retryCta'),
              onClick: () => void refetch(),
            }}
          />
        )}

        {!isLoading && !isError && (data?.presence.length ?? 0) === 0 && (
          <EmptyState
            icon={Users}
            title={t('flow.presence.noResults')}
            action={{
              label: t('flow.presence.refreshCta'),
              onClick: () => void refetch(),
            }}
          />
        )}

        {!isLoading && !isError && data && data.presence.length > 0 && (
          <ul className="space-y-1.5">
            {data.presence.map((entry: PresenceEntry) => {
              const fullName = `${entry.firstName ?? ''} ${entry.lastName ?? ''}`.trim();
              return (
                <li
                  key={entry.employeeId}
                  className="flex items-center gap-3 rounded-md p-2"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <Avatar initials={initials(fullName)} size="sm" />
                  <span
                    className="min-w-0 flex-1 text-body font-medium"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {fullName}
                  </span>
                  <span
                    className="rounded-sm px-1.5 py-0.5 text-meta font-semibold"
                    style={{ color: statusTone[entry.status] ?? 'var(--accent-fg)' }}
                  >
                    {statusLabel(entry.status)}
                  </span>
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </>
  );
};

export default PresenceScreen;
