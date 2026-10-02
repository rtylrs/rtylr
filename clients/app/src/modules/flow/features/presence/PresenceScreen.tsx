import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Users, AlertCircle, Building2, House, Shuffle, Palmtree, Circle } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { Avatar } from '@voxire/ui';
import { CardSkeleton } from '../../shared/components/Skeleton';
import type { PresenceEntry } from '../../shared/schemas/presence';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
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
 * A glyph per status, so the badge is not carrying its meaning in colour and
 * position alone (S-22). The text was already there, so this is scannability
 * rather than an accessibility fix.
 */
const statusGlyph: Record<string, React.ComponentType<{ size?: number }>> = {
  onsite: Building2,
  remote: House,
  hybrid: Shuffle,
  on_leave: Palmtree,
  not_started: Circle,
};

/**
 * Three groups, in the order someone scanning the screen cares about: who is
 * working, who is away, and who has not started. The flat list made a team of
 * any size a wall of names (S-22).
 */
const GROUPS = [
  { id: 'working', labelKey: 'flow.presence.working', statuses: ['onsite', 'remote', 'hybrid'] },
  { id: 'on_leave', labelKey: 'flow.presence.onLeave', statuses: ['on_leave'] },
  { id: 'not_started', labelKey: 'flow.presence.notStarted', statuses: ['not_started'] },
] as const;

/**
 * Business-wide presence, backed by /flow/presence/today - gated
 * server-side by view_directory (onion PR #267). Mode/status only, no
 * clocked/logged/gap minutes: a company-wide view watched by every
 * employee is a different exposure than the workspace-admin-gated Team view.
 */
const PresenceScreen = (): React.JSX.Element => {
  const { t } = useTranslation();
  const presenceQuery = usePresence();
  const { data, isLoading, refetch } = presenceQuery;
  // A 403 can retry into a paused query that never reaches isError, with
  // isLoading false and no data. Treating that as "no one here" told owners
  // the company was empty when the server had refused them.
  const isError = hasQueryFailed(presenceQuery);

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

        {!isLoading && !isError && data && data.presence.length === 0 && (
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
          <div className="space-y-5">
            {GROUPS.map((group) => {
              const entries = data.presence.filter((entry: PresenceEntry) =>
                (group.statuses as readonly string[]).includes(entry.status)
              );
              // A group with nobody in it is not news. An empty "On leave"
              // heading reads as a section that failed to load.
              if (entries.length === 0) return null;

              // The per-row badge only earns its place where the status varies
              // inside the group. Under "On leave" it repeated the heading
              // word for word on every row.
              const showRowStatus = group.statuses.length > 1;
              const GroupGlyph = statusGlyph[group.statuses[0]] ?? Circle;

              return (
                <section key={group.id}>
                  <h2
                    className="mb-1.5 flex items-center gap-1.5 text-flow-meta font-semibold uppercase tracking-wider"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <GroupGlyph size={12} />
                    {t(group.labelKey)}
                    <span className="font-normal tabular-nums">{entries.length}</span>
                  </h2>
                  <ul className="space-y-1.5">
                    {entries.map((entry: PresenceEntry) => {
                      const fullName = `${entry.firstName ?? ''} ${entry.lastName ?? ''}`.trim();
                      const Glyph = statusGlyph[entry.status] ?? Circle;
                      return (
                        <li
                          key={entry.employeeId}
                          className="flex items-center gap-3 rounded-md p-2"
                          style={{ background: 'var(--flow-hover)' }}
                        >
                          <Avatar initials={initials(fullName)} size="sm" />
                          <span
                            className="min-w-0 flex-1 text-flow-body font-medium"
                            style={{ color: 'var(--flow-text)' }}
                          >
                            {fullName}
                          </span>
                          {showRowStatus && (
                            <span
                              className="flex shrink-0 items-center gap-1 rounded-sm px-1.5 py-0.5 text-flow-meta font-semibold"
                              style={{ color: statusTone[entry.status] ?? 'var(--flow-text)' }}
                            >
                              <Glyph size={12} />
                              {statusLabel(entry.status)}
                            </span>
                          )}
                        </li>
                      );
                    })}
                  </ul>
                </section>
              );
            })}
          </div>
        )}
      </div>
    </>
  );
};

export default PresenceScreen;
