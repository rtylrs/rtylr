import React, { useState } from 'react';
import { useUserStore } from '@voxire/auth';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { Play, DollarSign, TrendingUp, Users, Clock } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { motion } from 'motion/react';
import { useListPayrollRuns, useTriggerPayrollRun } from '../../shared/hooks/queries/usePayroll';
import type {
  PayrollRunPayload,
  ListPayrollRunsParams,
} from '../../shared/hooks/queries/usePayroll';
import {
  Card,
  SectionHeader,
  Button,
  Skeleton,
  QueryError,
  Tabs,
  Pill,
} from '../../shared/components/ui';
import { useHrSettings } from '../../shared/hooks/useHrExtras';
import { formatCurrency } from '../../shared/lib/hrNumberUtils';
import { canHrPermission } from '../../shared/lib/hrPermissions';
import {
  PayrollConfigRequired,
  usePayrollConfigured,
} from '../../shared/components/PayrollConfigGuard';

const STATUS_COLORS: Record<string, string> = {
  draft: 'var(--hr-text-muted)',
  in_review: 'var(--hr-status-submitted)',
  approved: 'var(--hr-status-approved)',
  exported: 'var(--hr-status-paid)',
};

const FILTER_STATUSES = ['all', 'draft', 'in_review', 'approved', 'exported'] as const;
type StatusFilter = (typeof FILTER_STATUSES)[number];

const Payroll = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: settings } = useHrSettings();
  const role = useUserStore((state) => state.user?.role ?? state.role);
  const canRunPayroll = canHrPermission(role, 'hr:payroll:run');
  const payrollConfigured = usePayrollConfigured();
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [params] = useState<ListPayrollRunsParams>({ limit: 50, offset: 0 });

  const { data, isLoading, isError, isFetching, refetch } = useListPayrollRuns(params, {
    enabled: payrollConfigured,
  });
  const triggerRun = useTriggerPayrollRun(payrollConfigured);

  if (!payrollConfigured) return <PayrollConfigRequired />;

  const runs = data?.runs ?? [];
  const filteredRuns = runs.filter((r: PayrollRunPayload) =>
    statusFilter === 'all' ? true : r.status === statusFilter
  );

  const totalNet = runs.reduce((s: number, r: PayrollRunPayload) => s + (r.totalNet ?? 0), 0);
  const totalGross = runs.reduce((s: number, r: PayrollRunPayload) => s + (r.totalGross ?? 0), 0);
  const pendingCount = runs.filter((r: PayrollRunPayload) => r.status === 'in_review').length;

  const handleTriggerRun = (): void => {
    const now = new Date();
    const y = now.getFullYear();
    const m = now.getMonth();
    const start = `${y}-${String(m + 1).padStart(2, '0')}-01`;
    const lastDay = new Date(y, m + 1, 0).getDate();
    const end = `${y}-${String(m + 1).padStart(2, '0')}-${String(lastDay).padStart(2, '0')}`;
    triggerRun.mutate({ periodStart: start, periodEnd: end });
  };

  return (
    <>
      <HelmetContainer title="Payroll · rtylr" />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <SectionHeader
          title={t('payroll.title')}
          subtitle={t('payroll.subtitle')}
          right={
            canRunPayroll ? (
              <Button
                variant="primary"
                icon={<Play size={13} />}
                onClick={handleTriggerRun}
                disabled={triggerRun.isPending}
              >
                {t('payroll.triggerRun')}
              </Button>
            ) : undefined
          }
        />

        {/* Summary cards */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            gap: 12,
          }}
        >
          {[
            {
              label: t('payroll.stats.totalGross'),
              value: formatCurrency(totalGross, settings?.currency),
              icon: DollarSign,
              color: 'var(--accent-on-muted)',
              bg: 'var(--accent-muted)',
            },
            {
              label: t('payroll.stats.totalNet'),
              value: formatCurrency(totalNet, settings?.currency),
              icon: TrendingUp,
              color: 'var(--hr-status-approved)',
              bg: 'color-mix(in srgb, var(--hr-status-approved) 12%, transparent)',
            },
            {
              label: t('payroll.stats.pendingRuns'),
              value: String(pendingCount),
              icon: Clock,
              color: 'var(--hr-status-submitted)',
              bg: 'color-mix(in srgb, var(--hr-status-submitted) 12%, transparent)',
            },
          ].map((card) => (
            <motion.div
              key={card.label}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.2 }}
              style={{
                background: card.bg,
                border: `1px solid color-mix(in srgb, ${card.color} 30%, transparent)`,
                borderRadius: 12,
                padding: '16px 18px',
                display: 'flex',
                flexDirection: 'column',
                gap: 10,
              }}
            >
              <div
                style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}
              >
                <span style={{ fontSize: 12, fontWeight: 500, color: card.color }}>
                  {card.label}
                </span>
                <card.icon size={16} style={{ color: card.color }} />
              </div>
              <p style={{ margin: 0, fontSize: 22, fontWeight: 700, color: card.color }}>
                {card.value}
              </p>
            </motion.div>
          ))}
        </div>

        {/* Status filter tabs */}
        <Tabs
          tabs={FILTER_STATUSES.map((s) => t(`payroll.statuses.${s}`))}
          active={t(`payroll.statuses.${statusFilter}`)}
          onChange={(label) => {
            const idx = FILTER_STATUSES.findIndex((s) => t(`payroll.statuses.${s}`) === label);
            if (idx >= 0) setStatusFilter(FILTER_STATUSES[idx]);
          }}
        />

        {/* Runs list */}
        {isError ? (
          <QueryError onRetry={() => refetch()} isRetrying={isFetching} />
        ) : isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} height={72} />
            ))}
          </div>
        ) : filteredRuns.length === 0 ? (
          <EmptyState
            icon={DollarSign}
            title={t('payroll.emptyTitle')}
            description={t('payroll.emptyDescription')}
            action={
              statusFilter !== 'all'
                ? { label: t('common.clearFilters'), onClick: () => setStatusFilter('all') }
                : canRunPayroll
                  ? { label: t('payroll.triggerRun'), onClick: handleTriggerRun }
                  : { label: t('common.refresh'), onClick: () => refetch() }
            }
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {filteredRuns.map((run: PayrollRunPayload) => (
              <Link
                key={run.id}
                to="/payroll/$id"
                params={{ id: Number(run.id) }}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                <Card
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 16,
                    cursor: 'pointer',
                    transition: 'background 0.12s',
                  }}
                >
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div
                      style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}
                    >
                      <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                        {run.periodStart}, {run.periodEnd}
                      </span>
                      <Pill color={STATUS_COLORS[run.status ?? '']}>
                        {t(`payroll.statuses.${run.status}`)}
                      </Pill>
                    </div>
                    <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 3 }}>
                      <Users size={11} style={{ verticalAlign: 'middle', marginRight: 4 }} />
                      {t('payroll.employeeCount', { count: run.entries?.length ?? 0 })} ·{' '}
                      {t('payroll.netTotal')}:{' '}
                      {formatCurrency(run.totalNet ?? 0, settings?.currency)}
                    </div>
                  </div>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>
    </>
  );
};

export default Payroll;
