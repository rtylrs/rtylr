import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { serializeTaskFilters, type TaskFilters } from '../../../shared/filters/taskFilters';
import { priorityLabelKey } from '../../../shared/lib/priorityLabels';
import { PRIORITY_COLORS } from '../../../shared/stores/flowStore';
import type { HoursSummary, ProjectSummaryData, SummaryMetric } from '../projectSummary';
import { formatHours } from '../../../../insights/shared/lib/format';

/**
 * The cards on the project dashboard. Every number the summary computes is a
 * link to the project's List view, filtered through the shared filter
 * engine's URL `filter` param, so a lead can go from "3 overdue" to those
 * three tasks. The dashboard's own tiles and cards use the same pieces.
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

/** A titled dashboard card. */
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

const TILE_BOX: React.CSSProperties = {
  background: 'var(--flow-card)',
  border: '1px solid var(--flow-border)',
};

/**
 * One number with its label and an optional hint. With `link` the whole tile
 * opens the project's list filtered to the tasks it counts.
 */
export const StatTile: React.FC<{
  label: string;
  value: React.ReactNode;
  hint?: string;
  /** Colour of the hint; muted when absent. */
  tint?: string;
  link?: { projectId: string; filters: TaskFilters; 'aria-label'?: string };
  /** Marks the value with `data-summary-count`. */
  countKey?: string;
}> = ({ label, value, hint, tint, link, countKey }) => {
  const body = (
    <>
      <span
        className="block text-flow-meta font-semibold uppercase tracking-wider mb-1"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        {label}
      </span>
      <span className="flex items-baseline gap-2">
        <span
          className="text-flow-screen font-bold tabular-nums"
          style={{ color: 'var(--flow-text)' }}
          data-summary-count={countKey}
        >
          {value}
        </span>
        {hint && (
          <span className="text-flow-meta" style={{ color: tint ?? 'var(--flow-text-muted)' }}>
            {hint}
          </span>
        )}
      </span>
    </>
  );
  return link ? (
    <FilteredListLink
      projectId={link.projectId}
      filters={link.filters}
      aria-label={link['aria-label']}
      className="block rounded-md p-3 transition-colors hover:bg-[var(--flow-hover)]"
      style={TILE_BOX}
    >
      {body}
    </FilteredListLink>
  ) : (
    <div className="rounded-md p-3" style={TILE_BOX}>
      {body}
    </div>
  );
};

export interface BreakdownSegment {
  key: string;
  label: string;
  count: number;
  color: string;
  /** Makes the legend entry a link to these tasks. */
  filters?: TaskFilters;
}

/**
 * A stacked bar with a legend underneath: the status and priority breakdowns.
 * Segments with no tasks are left out of the bar but stay in the legend.
 */
export const BreakdownBar: React.FC<{
  projectId: string;
  segments: readonly BreakdownSegment[];
  /** Data attribute set on each legend count, e.g. `data-priority-count`. */
  countAttr?: string;
}> = ({ projectId, segments, countAttr }) => {
  const total = segments.reduce((sum, s) => sum + s.count, 0);
  return (
    <>
      <div
        className="flex h-6 rounded-md overflow-hidden"
        style={{ background: 'var(--flow-hover)' }}
        aria-hidden="true"
      >
        {segments.map((s) =>
          s.count > 0 ? (
            <div
              key={s.key}
              style={{ width: `${(s.count / total) * 100}%`, background: s.color }}
              title={`${s.label}: ${s.count}`}
            />
          ) : null
        )}
      </div>
      <div className="flex flex-wrap gap-2 mt-3 text-flow-meta">
        {segments.map((s) => {
          const content = (
            <>
              <span
                className="h-2 w-2 rounded-full"
                style={{ background: s.color }}
                aria-hidden="true"
              />
              {s.label}
              <span
                style={{ color: 'var(--flow-text-dim)' }}
                {...(countAttr ? { [countAttr]: s.key } : {})}
              >
                ({s.count})
              </span>
            </>
          );
          return s.filters ? (
            <FilteredListLink
              key={s.key}
              projectId={projectId}
              filters={s.filters}
              className="flex items-center gap-1.5 rounded-sm px-1.5 py-0.5 transition-colors hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {content}
            </FilteredListLink>
          ) : (
            <span
              key={s.key}
              className="flex items-center gap-1.5 px-1.5 py-0.5"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {content}
            </span>
          );
        })}
      </div>
    </>
  );
};

/**
 * Hours to one decimal with Latin digits, through the formatter the Insights
 * panel at the top of the dashboard uses, so both read the same.
 */
export const fmtHours = (hours: number, language: string): string =>
  formatHours(hours * 60, language);

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
        <StatTile
          key={tile.key}
          label={tile.label}
          value={tile.metric.count}
          hint={tile.hint}
          countKey={tile.key}
          link={{
            projectId,
            filters: tile.metric.filters,
            'aria-label': t('flow.projectDashboard.summary.openFiltered', {
              label: `${tile.label} (${tile.hint})`,
              n: tile.metric.count,
            }),
          }}
        />
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
        <BreakdownBar
          projectId={projectId}
          countAttr="data-priority-count"
          segments={priorities.map(({ priority, metric }) => ({
            key: priority,
            label: t(priorityLabelKey(priority)),
            count: metric.count,
            color: PRIORITY_COLORS[priority],
            filters: metric.filters,
          }))}
        />
      )}
    </SummaryCard>
  );
};
