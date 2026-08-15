import React, { useState } from 'react';
import { useUserStore } from '@voxire/auth';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { Play, DollarSign, TrendingUp, Users, Clock } from 'lucide-react';
import { motion } from 'motion/react';
import {
  useCreatePayrollAdjustment,
  useDeletePayrollAdjustment,
  useListPayrollAdjustments,
  useListPayrollRuns,
  useTriggerPayrollRun,
} from '../../shared/hooks/queries/usePayroll';
import { useListEmployees } from '../../shared/hooks/queries/useEmployees';
import type {
  PayrollRunPayload,
  ListPayrollRunsParams,
} from '../../shared/hooks/queries/usePayroll';
import {
  Card,
  SectionHeader,
  Button,
  EmptyState,
  Skeleton,
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

  const { data, isLoading } = useListPayrollRuns(params, { enabled: payrollConfigured });
  const triggerRun = useTriggerPayrollRun(payrollConfigured);
  const { data: adjustmentsData } = useListPayrollAdjustments(payrollConfigured && canRunPayroll);
  const { data: employeesData } = useListEmployees(
    { limit: 1000, offset: 0 },
    { enabled: payrollConfigured && canRunPayroll }
  );
  const createAdjustment = useCreatePayrollAdjustment();
  const deleteAdjustment = useDeletePayrollAdjustment();
  const [adjustmentEmployeeId, setAdjustmentEmployeeId] = useState('');
  const [adjustmentKind, setAdjustmentKind] = useState<'bonus' | 'deduction'>('bonus');
  const [adjustmentLabel, setAdjustmentLabel] = useState('');
  const [adjustmentAmount, setAdjustmentAmount] = useState('');
  const [adjustmentFrom, setAdjustmentFrom] = useState(() => new Date().toISOString().slice(0, 10));
  const [adjustmentTo, setAdjustmentTo] = useState('');

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
              color: 'var(--accent)',
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

        {canRunPayroll && (
          <Card>
            <SectionHeader
              title={t('payroll.adjustments.title')}
              subtitle={t('payroll.adjustments.subtitle')}
            />
            <form
              className="mt-4 grid gap-3 md:grid-cols-2 lg:grid-cols-4"
              onSubmit={(event) => {
                event.preventDefault();
                const employeeId = Number(adjustmentEmployeeId);
                const amount = Number(adjustmentAmount);
                if (!employeeId || !adjustmentLabel.trim() || !amount || !adjustmentFrom) return;
                createAdjustment.mutate(
                  {
                    employeeId,
                    kind: adjustmentKind,
                    label: adjustmentLabel.trim(),
                    amount,
                    effectiveFrom: adjustmentFrom,
                    effectiveTo: adjustmentTo || undefined,
                  },
                  {
                    onSuccess: () => {
                      setAdjustmentEmployeeId('');
                      setAdjustmentLabel('');
                      setAdjustmentAmount('');
                      setAdjustmentTo('');
                    },
                  }
                );
              }}
            >
              <select
                value={adjustmentEmployeeId}
                onChange={(event) => setAdjustmentEmployeeId(event.target.value)}
                required
                className="rounded-lg border bg-transparent px-3 py-2 text-sm"
              >
                <option value="">{t('payroll.adjustments.employeePlaceholder')}</option>
                {(employeesData?.employees ?? []).map((employee) => (
                  <option key={employee.id} value={employee.id}>
                    {employee.firstName} {employee.lastName} · #{employee.id}
                  </option>
                ))}
              </select>
              <select
                value={adjustmentKind}
                onChange={(event) => setAdjustmentKind(event.target.value as 'bonus' | 'deduction')}
                className="rounded-lg border bg-transparent px-3 py-2 text-sm"
              >
                <option value="bonus">{t('payroll.adjustments.bonus')}</option>
                <option value="deduction">{t('payroll.adjustments.deduction')}</option>
              </select>
              <input
                value={adjustmentLabel}
                onChange={(event) => setAdjustmentLabel(event.target.value)}
                placeholder={t('payroll.adjustments.labelPlaceholder')}
                required
                className="rounded-lg border bg-transparent px-3 py-2 text-sm"
              />
              <input
                type="number"
                min="0.01"
                step="0.01"
                value={adjustmentAmount}
                onChange={(event) => setAdjustmentAmount(event.target.value)}
                placeholder={t('payroll.adjustments.amountPlaceholder')}
                required
                className="rounded-lg border bg-transparent px-3 py-2 text-sm"
              />
              <label className="text-xs text-[var(--hr-text-muted)]">
                {t('payroll.adjustments.effectiveFrom')}
                <input
                  type="date"
                  value={adjustmentFrom}
                  onChange={(event) => setAdjustmentFrom(event.target.value)}
                  required
                  className="mt-1 w-full rounded-lg border bg-transparent px-3 py-2 text-sm"
                />
              </label>
              <label className="text-xs text-[var(--hr-text-muted)]">
                {t('payroll.adjustments.effectiveTo')}
                <input
                  type="date"
                  value={adjustmentTo}
                  onChange={(event) => setAdjustmentTo(event.target.value)}
                  className="mt-1 w-full rounded-lg border bg-transparent px-3 py-2 text-sm"
                />
              </label>
              <Button type="submit" variant="secondary" disabled={createAdjustment.isPending}>
                {createAdjustment.isPending
                  ? t('payroll.adjustments.saving')
                  : t('payroll.adjustments.add')}
              </Button>
            </form>
            {(adjustmentsData?.adjustments ?? []).length > 0 && (
              <div className="mt-4 divide-y" style={{ borderTop: '1px solid var(--hr-border)' }}>
                {adjustmentsData?.adjustments.map((adjustment) => (
                  <div
                    key={adjustment.id}
                    className="flex flex-wrap items-center justify-between gap-2 py-2 text-xs"
                  >
                    <span>
                      <strong>
                        {adjustment.kind === 'bonus' ? '+' : '-'}
                        {adjustment.amount.toFixed(2)}
                      </strong>{' '}
                      {adjustment.label} ·{' '}
                      {t('payroll.adjustments.employeeShort', { id: adjustment.employeeId })}
                    </span>
                    <button
                      type="button"
                      className="text-[var(--hr-error)]"
                      disabled={deleteAdjustment.isPending}
                      onClick={() => deleteAdjustment.mutate(adjustment.id)}
                    >
                      {t('common.delete')}
                    </button>
                  </div>
                ))}
              </div>
            )}
          </Card>
        )}

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
        {isLoading ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} height={72} />
            ))}
          </div>
        ) : filteredRuns.length === 0 ? (
          <EmptyState
            icon={<DollarSign size={20} />}
            title={t('payroll.emptyTitle')}
            description={t('payroll.emptyDescription')}
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
