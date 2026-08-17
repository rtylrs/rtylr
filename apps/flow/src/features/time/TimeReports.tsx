import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTimeSettingsStore } from '../../shared/stores/timeSettingsStore';
import { useWorkspaceMembers } from '../workspace-manager/useWorkspaceManager';
import GateNotice from '../../shared/components/GateNotice';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import { startOfWeek, endOfWeek } from '../../shared/lib/dateUtils';
import { todayLocalDate, useWorkDaysForWorkspace } from './useWorkDay';

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
  const { t } = useTranslation();
  const today = todayLocalDate();
  const weekStart = useTimeSettingsStore((s) => s.weekStart);
  const [from, setFrom] = useState(startOfWeek(today, weekStart));
  const [to, setTo] = useState(endOfWeek(today, weekStart));
  const activeWorkspaceId = useFlowStore((s) => s.activeWorkspaceId);

  useEffect(() => {
    setFrom(startOfWeek(today, weekStart));
    setTo(endOfWeek(today, weekStart));
  }, [today, weekStart]);

  const { data: members, isLoading: membersLoading } = useWorkspaceMembers(activeWorkspaceId);
  const { data, isLoading, isError } = useWorkDaysForWorkspace(activeWorkspaceId, from, to);

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
          <h1 className="text-[20px] font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.reports')}
          </h1>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <label htmlFor="reports-from" className="sr-only">
              {t('flow.time.reportsFrom')}
            </label>
            <input
              id="reports-from"
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              className="flow-focus min-h-[44px] rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
            <span className="text-[12px]" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.time.reportsRangeTo')}
            </span>
            <label htmlFor="reports-to" className="sr-only">
              {t('flow.time.reportsTo')}
            </label>
            <input
              id="reports-to"
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              className="flow-focus min-h-[44px] rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
          </div>
          <p className="mt-2 text-[11px]" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.reportsScopeNote')}
          </p>
        </header>

        {isError ? (
          <GateNotice reason={t('flow.time.teamAccessDenied')} className="text-[13px]" />
        ) : membersLoading || isLoading ? (
          <p className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.loadingDay')}
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-[13px]">
              <thead>
                <tr style={{ color: 'var(--flow-text-muted)' }}>
                  <th className="pb-2 text-start text-[10px] font-semibold uppercase tracking-wider">
                    {t('flow.time.reportsPerson')}
                  </th>
                  <th className="pb-2 text-end text-[10px] font-semibold uppercase tracking-wider">
                    {t('flow.time.clocked')}
                  </th>
                  <th className="pb-2 text-end text-[10px] font-semibold uppercase tracking-wider">
                    {t('flow.time.logged')}
                  </th>
                  <th className="pb-2 text-end text-[10px] font-semibold uppercase tracking-wider">
                    {t('flow.time.unaccounted')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {(members ?? []).map((member) => {
                  const totals = totalsByUserId.get(member.userId) ?? {
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
                      <td className="py-2 text-end tabular-nums">
                        {formatDurationClock(totals.clockedMinutes)}
                      </td>
                      <td className="py-2 text-end tabular-nums">
                        {formatDurationClock(totals.loggedMinutes)}
                      </td>
                      <td className="py-2 text-end tabular-nums">
                        {formatDurationClock(Math.abs(totals.gapMinutes))}
                      </td>
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
