import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { BarChart3 } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { DatePicker, EmptyState } from '@voxire/ui';
import { useNavigate } from '@tanstack/react-router';
import { useTimeSettingsStore } from '../../shared/stores/timeSettingsStore';
import { useFlowTeam } from '../../shared/hooks/queries/useFlowTeam';
import GateNotice from '../../shared/components/GateNotice';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import { startOfWeek, endOfWeek } from '../../shared/lib/dateUtils';
import { todayLocalDate, useTeamWorkDays } from './useWorkDay';

interface PersonTotals {
  clockedMinutes: number;
  loggedMinutes: number;
  gapMinutes: number;
  daysLogged: number;
}

/**
 * Totals per person across a range. Day-level only: WorkDay carries no
 * project or work-type breakdown, and there is no endpoint that lists lines
 * across a range - only per day. Grouping by project needs that endpoint,
 * which does not exist yet (docs/flow-time-tracking design, step 6). This
 * screen's data source swaps once it does; its layout does not have to.
 */
const TimeReports = (): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const today = todayLocalDate();
  const weekStart = useTimeSettingsStore((s) => s.weekStart);
  const trackAttendance = useTimeSettingsStore((s) => s.trackAttendance);
  const [from, setFrom] = useState(startOfWeek(today, weekStart));
  const [to, setTo] = useState(endOfWeek(today, weekStart));

  useEffect(() => {
    setFrom(startOfWeek(today, weekStart));
    setTo(endOfWeek(today, weekStart));
  }, [today, weekStart]);

  // Same stale TODO as TeamView: the roster was a hardcoded empty array, so this
  // screen could only ever render its empty state and the table below was
  // unreachable. useFlowTeam had already shipped. A member's `id` is their user
  // id, which is what the work-day totals are keyed on.
  const { data: allMembers = [], isLoading: membersLoading } = useFlowTeam();
  const members = useMemo(() => allMembers.filter((m) => m.hasFlowAccess), [allMembers]);
  const { data, isLoading, isError } = useTeamWorkDays(from, to);

  const totalsByUserId = useMemo(() => {
    const map = new Map<string, PersonTotals>();
    for (const day of data?.workDays ?? []) {
      if (!day.userId) continue;
      const totals = map.get(day.userId) ?? {
        clockedMinutes: 0,
        loggedMinutes: 0,
        gapMinutes: 0,
        daysLogged: 0,
      };
      totals.clockedMinutes += day.clockedMinutes ?? 0;
      totals.loggedMinutes += day.loggedMinutes ?? 0;
      totals.gapMinutes += day.gapMinutes ?? 0;
      totals.daysLogged += 1;
      map.set(day.userId, totals);
    }
    return map;
  }, [data]);

  return (
    <div className="h-full overflow-y-auto">
      <HelmetContainer title={`${t('flow.time.reports')} · Flow`} />
      <div className="mx-auto max-w-3xl px-4 py-6 sm:px-6">
        <header className="mb-6">
          <h1 className="text-screen font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.reports')}
          </h1>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <label htmlFor="reports-from" className="sr-only">
              {t('flow.time.reportsFrom')}
            </label>
            <DatePicker
              id="reports-from"
              value={from}
              onChange={setFrom}
              maxDate={to || undefined}
              locale={i18n.language}
              labels={{
                previousMonth: t('common.datePicker.previousMonth'),
                nextMonth: t('common.datePicker.nextMonth'),
              }}
              className="w-auto text-body"
            />
            <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.time.reportsRangeTo')}
            </span>
            <label htmlFor="reports-to" className="sr-only">
              {t('flow.time.reportsTo')}
            </label>
            <DatePicker
              id="reports-to"
              value={to}
              onChange={setTo}
              minDate={from || undefined}
              locale={i18n.language}
              labels={{
                previousMonth: t('common.datePicker.previousMonth'),
                nextMonth: t('common.datePicker.nextMonth'),
              }}
              className="w-auto text-body"
            />
          </div>
          <p className="mt-2 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.reportsScopeNote')}
          </p>
        </header>

        {isError ? (
          <GateNotice reason={t('flow.time.teamAccessDenied')} className="text-body" />
        ) : membersLoading || isLoading ? (
          <p className="text-body" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.loadingDay')}
          </p>
        ) : (members ?? []).length === 0 ? (
          <EmptyState
            icon={BarChart3}
            title={t('flow.time.reportsEmptyTitle')}
            description={t('flow.time.reportsEmptyDescription')}
            action={{
              label: t('flow.time.reportsEmptyCta'),
              onClick: (): void => {
                void navigate({ to: '/settings', search: { section: 'team' } });
              },
            }}
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-body">
              <thead>
                <tr style={{ color: 'var(--flow-text-muted)' }}>
                  <th className="pb-2 text-start text-meta font-semibold uppercase tracking-wider">
                    {t('flow.time.reportsPerson')}
                  </th>
                  {trackAttendance && (
                    <th className="pb-2 text-end text-meta font-semibold uppercase tracking-wider">
                      {t('flow.time.clocked')}
                    </th>
                  )}
                  <th className="pb-2 text-end text-meta font-semibold uppercase tracking-wider">
                    {t('flow.time.logged')}
                  </th>
                  {trackAttendance && (
                    <th className="pb-2 text-end text-meta font-semibold uppercase tracking-wider">
                      {t('flow.time.unaccounted')}
                    </th>
                  )}
                </tr>
              </thead>
              <tbody>
                {(members ?? []).map((member) => {
                  const totals = totalsByUserId.get(member.id) ?? {
                    clockedMinutes: 0,
                    loggedMinutes: 0,
                    gapMinutes: 0,
                    daysLogged: 0,
                  };
                  return (
                    <tr key={member.id} style={{ borderTop: '1px solid var(--flow-border)' }}>
                      <td className="py-2" style={{ color: 'var(--flow-text)' }}>
                        {member.name}
                      </td>
                      {trackAttendance && (
                        <td className="py-2 text-end tabular-nums">
                          {formatDurationClock(totals.clockedMinutes)}
                        </td>
                      )}
                      <td className="py-2 text-end tabular-nums">
                        {formatDurationClock(totals.loggedMinutes)}
                      </td>
                      {trackAttendance && (
                        <td className="py-2 text-end tabular-nums">
                          {formatDurationClock(Math.abs(totals.gapMinutes))}
                        </td>
                      )}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default TimeReports;
