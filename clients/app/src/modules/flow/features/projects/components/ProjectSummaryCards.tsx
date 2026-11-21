import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { serializeTaskFilters, type TaskFilters } from '../../../shared/filters/taskFilters';
import { PRIORITY_LABEL_KEY } from '../taskGroups';
import { PRIORITY_COLORS } from '../../../shared/stores/flowStore';
import type { HoursSummary, ProjectSummaryData, SummaryMetric } from '../projectSummary';

/**
 * The widgets T-591 adds to the project dashboard. Every number is a link to
 * the project's List view, filtered through the shared filter engine's URL
 * `filter` param, so a lead can go from "3 overdue" to those three tasks.
 */

/** Opens the project's list filtered to `filters`. */
export const FilteredListLink: React.FC<{
  projectId: string;
  filters: TaskFilters;
  className?: string;
  style?: React.CSSProperties;
  'aria-label'?: string;
  children: React.ReactNode;
}> = ({ projectId, filters, className, style, children, ...rest }) => (
  <Link
    to="/flow/p/$projectId/list"
    params={{ projectId }}
    search={{ filter: serializeTaskFilters(filters) }}
    className={['flow-focus', className].filter(Boolean).join(' ')}
    style={style}
    aria-label={rest['aria-label']}
    data-summary-link
  >
    {children}
  </Link>
);

export const SummaryCard: React.FC<{
  title: string;
  children: React.ReactNode;
  className?: string;
}> = ({ title, children, className }) => (
  <section
    className={['rounded-md p-4', className].filter(Boolean).join(' ')}
    style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
  >
    <h3
      className="text-flow-meta font-semibold uppercase tracking-wider mb-3"
      style={{ color: 'var(--flow-text-dim)' }}
    >
      {title}
    </h3>
    {children}
  </section>
);

const fmtHours = (value: number, locale: string): string =>
  new Intl.NumberFormat(locale, { maximumFractionDigits: 1 }).format(value);

/** Created, completed, updated in the last 7 days and due in the next 7. */
export const RecentChangeTiles: React.FC<{
  projectId: string;
  recent: ProjectSummaryData['recent'];
}> = ({ projectId, recent }) => {
  const { t } = useTranslation();
  const tiles: Array<{ key: string; label: string; hint: string; metric: SummaryMetric }> = [
    {
      key: 'created',
      label: t('flow.projectDashboard.summary.created'),
      hint: t('flow.projectDashboard.summary.last7Days'),
      metric: recent.created,
    },
    {
      key: 'completed',
      label: t('flow.projectDashboard.summary.completed'),
      hint: t('flow.projectDashboard.summary.last7Days'),
      metric: recent.completed,
    },
    {
      key: 'updated',
      label: t('flow.projectDashboard.summary.updated'),
      hint: t('flow.projectDashboard.summary.last7Days'),
      metric: recent.updated,
    },
    {
      key: 'dueSoon',
      label: t('flow.projectDashboard.summary.dueSoon'),
      hint: t('flow.projectDashboard.summary.next7Days'),
      metric: recent.dueSoon,
    },
  ];
  return (
    <section
      aria-label={t('flow.projectDashboard.summary.recentTitle')}
      className="grid grid-cols-2 md:grid-cols-4 gap-3"
    >
      {tiles.map((tile) => (
        <FilteredListLink
          key={tile.key}
          projectId={projectId}
          filters={tile.metric.filters}
          aria-label={t('flow.projectDashboard.summary.openFiltered', {
            label: `${tile.label} (${tile.hint})`,
            n: tile.metric.count,
          })}
          className="block rounded-md p-3 transition-colors hover:bg-[var(--flow-hover)]"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <span
            className="block text-flow-meta font-semibold uppercase tracking-wider mb-1"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            {tile.label}
          </span>
          <span className="flex items-baseline gap-2">
            <span
              className="text-flow-screen font-bold tabular-nums"
              style={{ color: 'var(--flow-text)' }}
              data-summary-count={tile.key}
            >
              {tile.metric.count}
            </span>
            <span className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {tile.hint}
            </span>
          </span>
        </FilteredListLink>
      ))}
    </section>
  );
};

/** Overdue work and active cycles with nobody on some of their tasks. */
export const NeedsAttentionCard: React.FC<{
  projectId: string;
  attention: ProjectSummaryData['attention'];
}> = ({ projectId, attention }) => {
  const { t } = useTranslation();
  const rows: Array<{ key: string; label: string; metric: SummaryMetric; tone: string }> = [
    {
      key: 'overdue',
      label: t('flow.projectDashboard.summary.overdue'),
      metric: attention.overdue,
      tone: 'var(--flow-danger-fg)',
    },
    ...attention.unassignedInCycles.map((gap) => ({
      key: `cycle:${gap.cycleId}`,
      label: t('flow.projectDashboard.summary.unassignedInCycle', { cycle: gap.cycleName }),
      metric: gap,
      tone: 'var(--flow-warn-fg)',
    })),
  ];
  const flagged = rows.filter((r) => r.metric.count > 0);
  return (
    <SummaryCard title={t('flow.projectDashboard.summary.needsAttention')}>
      {flagged.length === 0 ? (
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.projectDashboard.summary.allClear')}
        </p>
      ) : (
        <ul className="space-y-1">
          {flagged.map((row) => (
            <li key={row.key}>
              <FilteredListLink
                projectId={projectId}
                filters={row.metric.filters}
                className="flex items-center gap-3 rounded-sm px-2 py-1.5 transition-colors hover:bg-[var(--flow-hover)]"
              >
                <span
                  className="h-2 w-2 shrink-0 rounded-full"
                  style={{ background: row.tone }}
                  aria-hidden="true"
                />
                <span
                  className="flex-1 min-w-0 truncate text-flow-body"
                  style={{ color: 'var(--flow-text)' }}
                >
                  {row.label}
                </span>
                <span
                  className="text-flow-body font-semibold tabular-nums"
                  style={{ color: row.tone }}
                  data-attention-count={row.key}
                >
                  {row.metric.count}
                </span>
              </FilteredListLink>
            </li>
          ))}
        </ul>
      )}
    </SummaryCard>
  );
};

/** Estimated hours against the hours logged on those same tasks. */
export const HoursVsEstimateCard: React.FC<{ hours: HoursSummary }> = ({ hours }) => {
  const { t, i18n } = useTranslation();
  const locale = i18n.language;
  const pct = hours.estimated > 0 ? (hours.loggedOnEstimated / hours.estimated) * 100 : 0;
  const over = pct > 100;
  return (
    <SummaryCard title={t('flow.projectDashboard.summary.hoursTitle')}>
      <p className="text-flow-body" style={{ color: 'var(--flow-text)' }}>
        <span className="font-semibold tabular-nums" data-hours-logged>
          {t('flow.projectDashboard.summary.hoursValue', {
            hours: fmtHours(hours.loggedOnEstimated, locale),
          })}
        </span>{' '}
        <span style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.projectDashboard.summary.loggedOf', {
            hours: fmtHours(hours.estimated, locale),
          })}
        </span>
      </p>
      <div
        className="mt-2 h-1.5 rounded-full overflow-hidden"
        style={{ background: 'var(--flow-hover)' }}
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={Math.min(100, Math.round(pct))}
        aria-label={t('flow.projectDashboard.summary.hoursTitle')}
      >
        <div
          className="h-full rounded-full"
          style={{
            width: `${Math.min(100, pct)}%`,
            background: over ? 'var(--flow-danger-fg)' : 'var(--flow-status-approved)',
          }}
        />
      </div>
      <p className="mt-2 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.projectDashboard.summary.openEstimate', {
          hours: fmtHours(hours.openEstimated, locale),
        })}
        {hours.openUnestimated > 0 && (
          <>
            {' · '}
            {t('flow.projectDashboard.summary.unestimated', { n: hours.openUnestimated })}
          </>
        )}
      </p>
    </SummaryCard>
  );
};

/** Open tasks by priority: a bar and a legend whose entries open the list. */
export const PriorityBreakdownCard: React.FC<{
  projectId: string;
  priorities: ProjectSummaryData['priorities'];
}> = ({ projectId, priorities }) => {
  const { t } = useTranslation();
  const total = priorities.reduce((sum, p) => sum + p.metric.count, 0);
  return (
    <SummaryCard title={t('flow.projectDashboard.summary.priorityTitle')}>
      {total === 0 ? (
        <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.projectDashboard.summary.noOpenTasks')}
        </p>
      ) : (
        <>
          <div
            className="flex h-6 rounded-md overflow-hidden"
            style={{ background: 'var(--flow-hover)' }}
            aria-hidden="true"
          >
            {priorities.map(({ priority, metric }) =>
              metric.count > 0 ? (
                <div
                  key={priority}
                  style={{
                    width: `${(metric.count / total) * 100}%`,
                    background: PRIORITY_COLORS[priority],
                  }}
                />
              ) : null
            )}
          </div>
          <div className="flex flex-wrap gap-1 mt-3 text-flow-meta">
            {priorities.map(({ priority, metric }) => (
              <FilteredListLink
                key={priority}
                projectId={projectId}
                filters={metric.filters}
                className="flex items-center gap-1.5 rounded-sm px-1.5 py-0.5 transition-colors hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <span
                  className="h-2 w-2 rounded-full"
                  style={{ background: PRIORITY_COLORS[priority] }}
                  aria-hidden="true"
                />
                {t(PRIORITY_LABEL_KEY[priority])}
                <span style={{ color: 'var(--flow-text-dim)' }} data-priority-count={priority}>
                  ({metric.count})
                </span>
              </FilteredListLink>
            ))}
          </div>
        </>
      )}
    </SummaryCard>
  );
};
