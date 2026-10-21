import React, { useMemo, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ArrowRight, Phone, Plus, Upload } from 'lucide-react';
import { Button } from '../../shared/components/ui';
import { TierBadge } from '../../shared/components/crm';
import type { Tier } from '../../shared/lib/tiers';
import { useCrmDashboard } from '../../shared/hooks/queries/useCrmDashboard';
import {
  LoadFailed,
  RepFilterSelect,
  RepName,
  StatTile,
  useMoney,
  useRepNames,
} from '../sales/components/common';
import { TaskRow } from '../sales/components/Activity';
import {
  ACTIVITY_WINDOWS,
  type ActivityWindow,
  type LeadListFilters,
  useDealBoard,
  useDealStats,
  useLeadStats,
  useSalesActivity,
  useTaskCounts,
  useTaskList,
} from '../sales/data/queries';
import { DEFAULT_LEAD_FILTERS, NO_LEAD_FILTERS, useLeadFilters } from '../sales/data/leadFilters';
import { type LeadStatus, type RepFilter, WORKING_STATUSES } from '../sales/data/model';
import { ActivityChart } from './ActivityChart';
import { BarList } from './BarList';
import './dashboard.css';

/** Lead statuses in the order a lead moves through them. */
const FUNNEL: LeadStatus[] = ['new', 'attempting', 'connected', 'nurture', 'qualified'];
const TIERS: readonly string[] = ['Platinum', 'Gold', 'Silver', 'Bronze'];

const percentOf = (part: number, whole: number) =>
  whole > 0 ? Math.round((part * 100) / whole) : undefined;

const Card: React.FC<{
  title: string;
  summary?: React.ReactNode;
  link?: { label: string; onClick: () => void };
  className?: string;
  flush?: boolean;
  refreshing?: boolean;
  children: React.ReactNode;
}> = ({ title, summary, link, className, flush, refreshing, children }) => (
  <section className={`crm-dash-card ${className ?? ''}`} aria-label={title}>
    <div className="crm-dash-card-head">
      <div>
        <h2>{title}</h2>
        {summary && <p>{summary}</p>}
      </div>
      {link && (
        <button type="button" className="crm-dash-card-link" onClick={link.onClick}>
          {link.label}
          <ArrowRight size={13} />
        </button>
      )}
    </div>
    <div
      className={`crm-dash-card-body crm-dash-loading${flush ? ' is-flush' : ''}${refreshing ? ' is-refreshing' : ''}`}
    >
      {children}
    </div>
  </section>
);

const Pending: React.FC = () => {
  const { t } = useTranslation();
  return (
    <p className="crm-dash-empty" role="status">
      {t('crmSales.common.loading')}
    </p>
  );
};

const Dashboard: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const money = useMoney();
  const names = useRepNames();
  const setLeadFilters = useLeadFilters((state) => state.setFilters);
  const [days, setDays] = useState<ActivityWindow>(30);
  const [rep, setRep] = useState<RepFilter>('all');
  const number = (value?: number) => new Intl.NumberFormat(i18n.language).format(value ?? 0);
  const period = t(`crm.dashboard.period.${days}`);

  const leadFilters = useMemo<LeadListFilters>(
    () => ({
      ...DEFAULT_LEAD_FILTERS,
      reps: rep === 'all' ? [] : [rep === 'unassigned' ? '0' : rep],
    }),
    [rep]
  );
  const leadStats = useLeadStats(leadFilters);
  const dealStats = useDealStats(undefined, rep);
  const board = useDealBoard({ rep, search: '' });
  const activity = useSalesActivity(days, rep);
  const taskCounts = useTaskCounts(rep);
  const dueTasks = useTaskList({ scope: 'open', rep, due: 'due' }, 6);
  const customers = useCrmDashboard();

  /** Opens the leads list on these filters, keeping the dashboard's rep. */
  const openLeads = (change: Partial<LeadListFilters>) => {
    setLeadFilters({ ...NO_LEAD_FILTERS, reps: leadFilters.reps, search: '', ...change });
    void navigate({ to: '/crm/sales/leads' });
  };

  const byStatus = new Map(
    (leadStats.data?.byStatus ?? []).map((item) => [item.status ?? '', item.count ?? 0])
  );
  const totalLeads = [...byStatus.values()].reduce((sum, value) => sum + value, 0);
  const overdue = leadStats.data?.followUpsOverdue ?? 0;
  const due = overdue + (leadStats.data?.followUpsToday ?? 0);

  const totals = activity.data;
  const calls = totals?.calls ?? 0;
  const reachedShare = percentOf(totals?.reached ?? 0, calls);
  const won = totals?.dealsWon ?? 0;
  const lost = totals?.dealsLost ?? 0;
  const winRate = percentOf(won, won + lost);

  const openStages = (board.data?.columns ?? []).filter((column) => column.stage?.kind === 'open');
  const anyValue = openStages.some((column) => Number(column.amount ?? 0) > 0);
  const repRows = totals?.reps ?? [];
  const topCustomers = customers.data?.topCustomers ?? [];
  const tasks = dueTasks.data?.pages.flatMap((page) => page.tasks ?? []) ?? [];

  return (
    <div className="sales-page crm-dash">
      <HelmetContainer title={`${t('crm.dashboard.pageTitle')} · rtylr`} />
      <header className="sales-header">
        <div>
          <h1>{t('crm.dashboard.pageTitle')}</h1>
          <p>{t('crm.dashboard.subtitle')}</p>
        </div>
        <div className="crm-dash-controls">
          <div
            className="sales-segmented"
            role="radiogroup"
            aria-label={t('crm.dashboard.windowLabel')}
          >
            {ACTIVITY_WINDOWS.map((window) => (
              <button
                key={window}
                type="button"
                role="radio"
                aria-checked={days === window}
                onClick={() => setDays(window)}
              >
                {t(`crm.dashboard.window.${window}`)}
              </button>
            ))}
          </div>
          <RepFilterSelect value={rep} onChange={setRep} label={t('crm.dashboard.repLabel')} />
        </div>
      </header>

      <div className="sales-stats">
        <StatTile
          label={t('crm.dashboard.stats.pipeline')}
          value={dealStats.data ? money(dealStats.data.openAmount) : '—'}
          hint={
            dealStats.data
              ? t('crm.dashboard.stats.pipelineHint', { count: dealStats.data.openDeals ?? 0 })
              : undefined
          }
          onClick={() => void navigate({ to: '/crm/sales/deals' })}
        />
        <StatTile
          label={t('crm.dashboard.stats.won')}
          value={totals ? money(totals.wonAmount) : '—'}
          hint={totals ? t('crm.dashboard.stats.wonHint', { count: won }) : undefined}
        />
        <StatTile
          label={t('crm.dashboard.stats.winRate')}
          value={winRate === undefined ? '—' : `${number(winRate)}%`}
          hint={
            totals
              ? won + lost > 0
                ? t('crm.dashboard.stats.winRateHint', { won: number(won), lost: number(lost) })
                : t('crm.dashboard.stats.noneClosed')
              : undefined
          }
        />
        <StatTile
          label={t('crm.dashboard.stats.calls')}
          value={totals ? number(calls) : '—'}
          hint={
            totals
              ? reachedShare === undefined
                ? undefined
                : t('crm.dashboard.stats.callsHint', { percent: number(reachedShare) })
              : undefined
          }
        />
        <StatTile
          label={t('crm.dashboard.stats.followUps')}
          value={leadStats.data ? number(due) : '—'}
          hint={
            overdue > 0 ? t('crm.dashboard.stats.followUpsHint', { count: overdue }) : undefined
          }
          warning={overdue > 0}
          onClick={() => openLeads({ statuses: WORKING_STATUSES, followUp: 'due' })}
        />
        <StatTile
          label={t('crm.dashboard.stats.customers')}
          value={customers.data ? number(customers.data.totalCustomers) : '—'}
          hint={
            customers.data
              ? t('crm.dashboard.stats.customersHint', {
                  count: customers.data.newThisMonth ?? 0,
                  value: number(customers.data.newThisMonth),
                })
              : undefined
          }
          onClick={() => void navigate({ to: '/crm/customers' })}
        />
      </div>

      <div className="crm-dash-grid">
        <Card
          className="crm-dash-wide"
          title={t('crm.dashboard.calls.title')}
          summary={
            totals && calls > 0
              ? t('crm.dashboard.calls.summary', {
                  count: calls,
                  calls: number(calls),
                  percent: number(reachedShare ?? 0),
                  period,
                })
              : period
          }
          link={{
            label: t('crm.dashboard.calls.startCalling'),
            onClick: () => void navigate({ to: '/crm/sales/leads/call' }),
          }}
          refreshing={activity.isPlaceholderData}
        >
          {activity.isError && !totals ? (
            <LoadFailed
              title={t('crm.dashboard.loadFailedTitle')}
              onRetry={() => void activity.refetch()}
            />
          ) : !totals ? (
            <Pending />
          ) : calls === 0 ? (
            <div className="crm-dash-empty">
              <span>{t('crm.dashboard.calls.empty', { period })}</span>
              <Button onClick={() => void navigate({ to: '/crm/sales/leads/call' })}>
                <Phone size={13} />
                {t('crm.dashboard.calls.startCalling')}
              </Button>
            </div>
          ) : (
            <ActivityChart days={totals.days ?? []} />
          )}
        </Card>

        <Card
          className="crm-dash-narrow"
          title={t('crm.dashboard.leads.title')}
          summary={
            leadStats.data ? t('crm.dashboard.leads.summary', { count: totalLeads }) : undefined
          }
          link={{ label: t('crm.dashboard.leads.all'), onClick: () => openLeads({}) }}
          refreshing={leadStats.isPlaceholderData}
        >
          {leadStats.isError && !leadStats.data ? (
            <LoadFailed
              title={t('crm.dashboard.loadFailedTitle')}
              onRetry={() => void leadStats.refetch()}
            />
          ) : !leadStats.data ? (
            <Pending />
          ) : totalLeads === 0 ? (
            <div className="crm-dash-empty">
              <span>{t('crm.dashboard.leads.empty')}</span>
              <div className="crm-dash-empty-actions">
                <Button
                  variant="outline"
                  onClick={() => void navigate({ to: '/crm/sales/leads/import' })}
                >
                  <Upload size={13} />
                  {t('crm.dashboard.leads.import')}
                </Button>
                <Button onClick={() => openLeads({})}>
                  <Plus size={13} />
                  {t('crm.dashboard.leads.add')}
                </Button>
              </div>
            </div>
          ) : (
            <>
              <BarList
                label={t('crm.dashboard.leads.title')}
                items={FUNNEL.map((status) => ({
                  id: status,
                  label: t(`crmSales.leadStatus.${status}`),
                  value: byStatus.get(status) ?? 0,
                  display: number(byStatus.get(status) ?? 0),
                  onClick: () => openLeads({ statuses: [status] }),
                }))}
              />
              {totals && (
                <div className="crm-dash-foot">
                  <span>
                    {t('crm.dashboard.leads.added', { count: totals.leadsAdded ?? 0 })}{' '}
                    <strong>{number(totals.leadsAdded)}</strong>
                  </span>
                  <span>
                    {t('crm.dashboard.leads.qualified', { count: totals.leadsQualified ?? 0 })}{' '}
                    <strong>{number(totals.leadsQualified)}</strong>
                  </span>
                </div>
              )}
            </>
          )}
        </Card>

        <Card
          className="crm-dash-half"
          title={board.data?.pipeline?.name || t('crm.dashboard.pipeline.title')}
          summary={
            dealStats.data
              ? t('crm.dashboard.pipeline.summary', {
                  count: dealStats.data.openDeals ?? 0,
                  weighted: money(dealStats.data.weightedAmount),
                })
              : undefined
          }
          link={{
            label: t('crm.dashboard.pipeline.all'),
            onClick: () => void navigate({ to: '/crm/sales/deals' }),
          }}
          refreshing={board.isPlaceholderData}
        >
          {board.isError && !board.data ? (
            <LoadFailed
              title={t('crm.dashboard.loadFailedTitle')}
              onRetry={() => void board.refetch()}
            />
          ) : !board.data ? (
            <Pending />
          ) : openStages.every((column) => (column.deals ?? 0) === 0) ? (
            <div className="crm-dash-empty">
              <span>{t('crm.dashboard.pipeline.empty')}</span>
              <Button variant="outline" onClick={() => openLeads({ statuses: WORKING_STATUSES })}>
                {t('crm.dashboard.pipeline.workLeads')}
              </Button>
            </div>
          ) : (
            <>
              <BarList
                label={t('crm.dashboard.pipeline.title')}
                items={openStages.map((column) => ({
                  id: column.stage?.id ?? '',
                  label: column.stage?.name ?? '',
                  value: anyValue ? Number(column.amount ?? 0) : (column.deals ?? 0),
                  display: anyValue ? money(column.amount) : number(column.deals),
                  hint: anyValue
                    ? t('crm.dashboard.pipeline.deals', { count: column.deals ?? 0 })
                    : undefined,
                  onClick: () => void navigate({ to: '/crm/sales/deals' }),
                }))}
              />
              {dealStats.data && (dealStats.data.averageDaysToWin ?? 0) > 0 && (
                <div className="crm-dash-foot">
                  <span>
                    {t('crm.dashboard.pipeline.averageWon')}{' '}
                    <strong>{money(dealStats.data.averageWonAmount)}</strong>
                  </span>
                  <span>
                    {t('crm.dashboard.pipeline.daysToWin', {
                      count: Math.round(dealStats.data.averageDaysToWin ?? 0),
                    })}
                  </span>
                </div>
              )}
            </>
          )}
        </Card>

        <Card
          className="crm-dash-half crm-dash-due"
          title={t('crm.dashboard.tasks.title')}
          summary={
            taskCounts.data
              ? t('crm.dashboard.tasks.summary', {
                  overdue: number(taskCounts.data.overdue),
                  today: number(taskCounts.data.dueToday),
                })
              : undefined
          }
          link={{
            label: t('crm.dashboard.tasks.all'),
            onClick: () => void navigate({ to: '/crm/sales/tasks' }),
          }}
          flush
        >
          {dueTasks.isError && tasks.length === 0 ? (
            <LoadFailed
              title={t('crm.dashboard.loadFailedTitle')}
              onRetry={() => void dueTasks.refetch()}
            />
          ) : dueTasks.isPending ? (
            <Pending />
          ) : tasks.length === 0 ? (
            <p className="crm-dash-empty">{t('crm.dashboard.tasks.empty')}</p>
          ) : (
            tasks
              .slice(0, 6)
              .map((task) => (
                <TaskRow
                  key={task.id}
                  task={task}
                  showSubject
                  repName={names.get(task.assignedTo ?? '')}
                />
              ))
          )}
        </Card>

        <Card
          className="crm-dash-wide"
          title={t('crm.dashboard.team.title')}
          summary={period}
          flush
          refreshing={activity.isPlaceholderData}
        >
          {!totals ? (
            activity.isError ? (
              <LoadFailed
                title={t('crm.dashboard.loadFailedTitle')}
                onRetry={() => void activity.refetch()}
              />
            ) : (
              <Pending />
            )
          ) : repRows.length === 0 ? (
            <p className="crm-dash-empty">{t('crm.dashboard.team.empty', { period })}</p>
          ) : (
            <div className="sales-table-wrap">
              <table className="sales-table crm-dash-table">
                <thead>
                  <tr>
                    <th>{t('crm.dashboard.team.rep')}</th>
                    <th className="is-number">{t('crm.dashboard.team.calls')}</th>
                    <th className="is-number">{t('crm.dashboard.team.reached')}</th>
                    <th className="is-number">{t('crm.dashboard.team.qualified')}</th>
                    <th className="is-number">{t('crm.dashboard.team.won')}</th>
                    <th className="is-number">{t('crm.dashboard.team.wonValue')}</th>
                  </tr>
                </thead>
                <tbody>
                  {repRows.map((row) => {
                    const share = percentOf(row.reached ?? 0, row.calls ?? 0);
                    return (
                      <tr key={row.repId}>
                        <td>
                          <RepName id={row.repId === '0' ? undefined : row.repId} names={names} />
                        </td>
                        <td className="is-number">{number(row.calls)}</td>
                        <td className="is-number">
                          {share === undefined ? '—' : `${number(share)}%`}
                        </td>
                        <td className="is-number">{number(row.leadsQualified)}</td>
                        <td className="is-number">{number(row.dealsWon)}</td>
                        <td className="is-number">{money(row.wonAmount)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </Card>

        <Card
          className="crm-dash-narrow"
          title={t('crm.dashboard.customers.title')}
          summary={
            customers.data
              ? t('crm.dashboard.customers.summary', {
                  count: customers.data.activeLoyaltyMembers ?? 0,
                  value: number(customers.data.activeLoyaltyMembers),
                })
              : undefined
          }
          link={{
            label: t('crm.dashboard.customers.all'),
            onClick: () => void navigate({ to: '/crm/customers' }),
          }}
          flush
        >
          {customers.isError && !customers.data ? (
            <LoadFailed
              title={t('crm.dashboard.loadFailedTitle')}
              onRetry={() => void customers.refetch()}
            />
          ) : !customers.data ? (
            <Pending />
          ) : topCustomers.length === 0 ? (
            <p className="crm-dash-empty">{t('crm.dashboard.customers.empty')}</p>
          ) : (
            <ol className="crm-dash-people">
              {topCustomers.slice(0, 6).map((customer, index) => (
                <li key={customer.id ?? `${customer.name}-${index}`}>
                  <button
                    type="button"
                    className="crm-dash-person"
                    disabled={!customer.id}
                    onClick={() => void navigate({ to: `/crm/customers/${customer.id}` })}
                  >
                    <span className="sales-avatar" aria-hidden="true">
                      {customer.name
                        .split(' ')
                        .map((part) => part[0])
                        .join('')
                        .slice(0, 2)
                        .toUpperCase()}
                    </span>
                    <span className="crm-dash-person-name">{customer.name}</span>
                    {TIERS.includes(customer.tier) && <TierBadge tier={customer.tier as Tier} />}
                    <span className="crm-dash-person-value">{money(customer.spend)}</span>
                  </button>
                </li>
              ))}
            </ol>
          )}
        </Card>
      </div>
    </div>
  );
};

export default Dashboard;
