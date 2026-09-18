import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Users } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { DatePicker, EmptyState } from '@voxire/ui';
import { useNavigate } from '@tanstack/react-router';
import { useFlowTeam } from '../../shared/hooks/queries/useFlowTeam';
import GateNotice from '../../shared/components/GateNotice';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import { todayLocalDate, useTeamWorkDays, type WorkDay } from './useWorkDay';

const labelClass = 'mb-1 block text-meta font-semibold uppercase tracking-wider';

/**
 * The team view: one row per workspace member, for the day selected.
 *
 * Requires at least admin on the workspace, enforced server-side by the same
 * call this screen makes - a member without it gets a real 403, rendered as
 * teamAccessDenied rather than a raw error or an empty roster that looks like
 * nobody logged anything.
 */
const TeamView = (): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [date, setDate] = useState(todayLocalDate());

  // The roster was a hardcoded empty array, so the empty state below was the
  // only branch this screen could ever render and the member table underneath
  // it was unreachable. useFlowTeam already shipped, so the TODO it was waiting
  // on was stale.
  //
  // It returns the whole business team; the people who can appear on a Flow
  // time sheet are the ones with Flow access, which is the same split
  // FlowSettings makes. A member's `id` is their user id, which is what
  // workDays is keyed on.
  const { data: allMembers = [], isLoading: membersLoading } = useFlowTeam();
  const members = useMemo(() => allMembers.filter((m) => m.apps?.includes('flow')), [allMembers]);
  const { data, isLoading, isError } = useTeamWorkDays(date, date);

  const dayByUserId = useMemo(() => {
    const map = new Map<string, WorkDay>();
    for (const day of data?.workDays ?? []) {
      if (day.userId) map.set(day.userId, day);
    }
    return map;
  }, [data]);

  return (
    <div className="h-full overflow-y-auto">
      <HelmetContainer title={`${t('flow.time.team')} · Flow`} />
      <div className="mx-auto max-w-3xl px-4 py-6 sm:px-6">
        <header className="mb-6">
          <h1 className="text-screen font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.team')}
          </h1>
          <label htmlFor="team-view-date" className="sr-only">
            {t('flow.time.date')}
          </label>
          <DatePicker
            id="team-view-date"
            value={date}
            onChange={setDate}
            locale={i18n.language}
            wrapperClassName="mt-2"
            className="w-auto"
            labels={{
              previousMonth: t('common.datePicker.previousMonth'),
              nextMonth: t('common.datePicker.nextMonth'),
            }}
          />
        </header>

        {isError ? (
          <GateNotice reason={t('flow.time.teamAccessDenied')} className="text-body" />
        ) : membersLoading || isLoading ? (
          <p className="text-body" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.loadingDay')}
          </p>
        ) : (members ?? []).length === 0 ? (
          <EmptyState
            icon={Users}
            title={t('flow.time.teamEmptyTitle')}
            description={t('flow.time.teamEmptyDescription')}
            action={{
              label: t('flow.time.teamEmptyCta'),
              onClick: (): void => {
                void navigate({ to: '/settings', search: { section: 'team' } });
              },
            }}
          />
        ) : (
          <ul className="space-y-1.5">
            {(members ?? []).map((member) => {
              const day = dayByUserId.get(member.id);
              return (
                <li
                  key={member.id}
                  className="flex items-center justify-between gap-3 rounded-md p-3"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <span className="text-body font-medium" style={{ color: 'var(--flow-text)' }}>
                    {member.name}
                  </span>
                  {day ? (
                    <div className="flex items-baseline gap-4">
                      <Figure
                        label={t('flow.time.logged')}
                        value={formatDurationClock(day.loggedMinutes ?? 0)}
                      />
                      <Figure
                        label={t('flow.time.unaccounted')}
                        value={formatDurationClock(Math.abs(day.gapMinutes ?? 0))}
                        tone={
                          (day.gapMinutes ?? 0) === 0 ? undefined : 'var(--flow-status-pending)'
                        }
                      />
                    </div>
                  ) : (
                    <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                      {t('flow.time.teamNotStarted')}
                    </span>
                  )}
                </li>
              );
            })}
          </ul>
        )}
      </div>
    </div>
  );
};

const Figure = ({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone?: string;
}): React.JSX.Element => (
  <div className="text-right">
    <span className={labelClass} style={{ color: 'var(--flow-text-muted)' }}>
      {label}
    </span>
    <span
      className="text-body font-semibold tabular-nums"
      style={{ color: tone ?? 'var(--flow-text)' }}
    >
      {value}
    </span>
  </div>
);

export default TeamView;
