import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Users } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { DatePicker } from '@voxire/ui';
import { useNavigate } from '@tanstack/react-router';
import { useFlowTeam } from '../../shared/hooks/queries/useFlowTeam';
import GateNotice from '../../shared/components/GateNotice';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import { todayLocalDate, useTeamWorkDays, type WorkDay } from './useWorkDay';

const labelClass = 'mb-1 block text-flow-meta font-semibold uppercase tracking-wider';

const TeamView = (): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const [date, setDate] = useState(todayLocalDate());

  const membersQuery = useFlowTeam();
  const { data: allMembers = [], isLoading: membersLoading } = membersQuery;
  const members = useMemo(() => allMembers.filter((m) => m.hasFlowAccess), [allMembers]);
  const daysQuery = useTeamWorkDays(date, date);
  const { data, isLoading, isError, error } = daysQuery;
  const accessDenied =
    typeof error === 'object' &&
    error !== null &&
    (('statusCode' in error && error.statusCode === 403) ||
      ('status' in error && error.status === 403));

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
          <h1 className="text-flow-screen font-bold" style={{ color: 'var(--flow-text)' }}>
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
              previousMonth: t('flowCommon.datePicker.previousMonth'),
              nextMonth: t('flowCommon.datePicker.nextMonth'),
            }}
          />
        </header>

        {isError || membersQuery.isError ? (
          accessDenied ? (
            <GateNotice reason={t('flow.time.teamAccessDenied')} className="text-flow-body" />
          ) : (
            <ListLoadFailed
              onRetry={() => {
                void daysQuery.refetch();
                void membersQuery.refetch();
              }}
            />
          )
        ) : membersLoading || isLoading ? (
          <p className="text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
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
                void navigate({ to: '/team' });
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
                  <span
                    className="text-flow-body font-medium"
                    style={{ color: 'var(--flow-text)' }}
                  >
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
                        value={`${(day.gapMinutes ?? 0) < 0 ? '−' : ''}${formatDurationClock(Math.abs(day.gapMinutes ?? 0))}`}
                        tone={
                          (day.gapMinutes ?? 0) === 0 ? undefined : 'var(--flow-status-pending)'
                        }
                      />
                    </div>
                  ) : (
                    <span className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
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
      className="text-flow-body font-semibold tabular-nums"
      style={{ color: tone ?? 'var(--flow-text)' }}
    >
      {value}
    </span>
  </div>
);

export default TeamView;
