import React, { useRef, useCallback } from 'react';
import { getRouteApi, Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { CheckCircle, Download, ArrowLeft, AlertTriangle, DollarSign } from 'lucide-react';
import { useGetPayrollRun, useApprovePayrollRun } from '../../shared/hooks/queries/usePayroll';
import type { PayrollEntryPayload } from '../../shared/hooks/queries/usePayroll';
import { ErrorBoundary } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import {
  SectionHeader,
  Button,
  Skeleton,
  Pill,
  Breadcrumbs,
  EmptyState,
} from '../../shared/components/ui';

const STATUS_COLORS: Record<string, string> = {
  draft: 'var(--hr-text-muted)',
  in_review: 'var(--hr-status-submitted)',
  approved: 'var(--hr-status-approved)',
  exported: 'var(--hr-status-paid)',
};

function fmt(n: number): string {
  return n.toLocaleString('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
  });
}

const ROW_HEIGHT = 56;
const HEADER_HEIGHT = 44;
const COLUMNS = '2fr 1fr 1fr 1fr 1fr 1fr 1fr';
const MIN_WIDTH = 860;

// Route id, not path: `hrConfigRoute` is pathless with `id: 'hr-config'`, so
// every feature route is registered beneath it.
const routeApi = getRouteApi('/hr-config/payroll/$id');

const PayrollRunDetailContent = (): React.JSX.Element => {
  const { id: runId } = routeApi.useParams();
  const { t } = useTranslation();
  const { data: run, isLoading } = useGetPayrollRun(runId);
  const approveRun = useApprovePayrollRun();
  const parentRef = useRef<HTMLDivElement>(null);

  const entries = run?.entries ?? [];

  const virtualizer = useVirtualizer({
    count: entries.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  });

  const handleApprove = useCallback((): void => {
    if (runId) approveRun.mutate({ runId });
  }, [runId, approveRun]);

  const handleExport = useCallback((): void => {
    if (!run) return;
    const headers = [
      'Employee ID',
      'Hours Worked',
      'Base Pay',
      'Overtime Pay',
      'Deductions',
      'Bonuses',
      'Net Pay',
    ];
    const rows = (run.entries ?? []).map((e: PayrollEntryPayload) => [
      e.employeeId,
      e.hoursWorked,
      e.basePay,
      e.overtimePay,
      e.deductions,
      e.bonuses,
      e.netPay,
    ]);
    const csv = [headers, ...rows].map((r) => r.join(',')).join('\n');
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `payroll-run-${run.id}-${run.periodStart}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  }, [run]);

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={400} />
      </div>
    );
  }

  // Either the param failed `payrollRunParamsSchema` (so the query never ran)
  // or the backend has no run with this id. Both are the same user-facing
  // condition: this URL does not name a run.
  if (!run) {
    return (
      <EmptyState
        icon={<DollarSign size={20} />}
        title={t('payroll.runDetail.notFound')}
        description={t('payroll.runDetail.notFoundDescription')}
        action={
          <Link
            to="/payroll"
            style={{ color: 'var(--hr-accent)', textDecoration: 'none', fontSize: 12 }}
          >
            {t('payroll.backToRuns')}
          </Link>
        }
      />
    );
  }

  const totalDeductions = entries.reduce(
    (s: number, e: PayrollEntryPayload) => s + (e.deductions ?? 0),
    0
  );

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <Breadcrumbs
        items={[
          { label: t('payroll.title'), href: '/payroll' },
          { label: `${run.periodStart}, ${run.periodEnd}` },
        ]}
      />

      <SectionHeader
        title={`${run.periodStart}, ${run.periodEnd}`}
        subtitle={t('payroll.runDetail.subtitle', { count: entries.length })}
        right={
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <Pill color={STATUS_COLORS[run.status ?? '']}>
              {t(`payroll.statuses.${run.status}`)}
            </Pill>
            {run.status === 'in_review' && (
              <Button
                variant="primary"
                icon={<CheckCircle size={13} />}
                onClick={handleApprove}
                disabled={approveRun.isPending}
              >
                {t('payroll.runDetail.approve')}
              </Button>
            )}
            <Button variant="secondary" icon={<Download size={13} />} onClick={handleExport}>
              {t('payroll.runDetail.exportCsv')}
            </Button>
          </div>
        }
      />

      {/* Summary row */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
          gap: 12,
        }}
      >
        {[
          {
            label: t('payroll.stats.totalGross'),
            value: fmt(run.totalGross ?? 0),
            color: 'var(--accent)',
          },
          {
            label: t('payroll.stats.totalDeductions'),
            value: fmt(totalDeductions),
            color: 'var(--hr-error)',
          },
          {
            label: t('payroll.stats.totalNet'),
            value: fmt(run.totalNet ?? 0),
            color: 'var(--hr-status-approved)',
          },
        ].map((stat) => (
          <div
            key={stat.label}
            style={{
              background: `color-mix(in srgb, ${stat.color} 12%, transparent)`,
              border: `1px solid color-mix(in srgb, ${stat.color} 30%, transparent)`,
              borderRadius: 10,
              padding: '12px 14px',
            }}
          >
            <div
              style={{
                fontSize: 10,
                fontWeight: 600,
                color: stat.color,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
              }}
            >
              {stat.label}
            </div>
            <div style={{ fontSize: 18, fontWeight: 700, color: stat.color, marginTop: 4 }}>
              {stat.value}
            </div>
          </div>
        ))}
      </div>

      {/* Virtualised table */}
      <div
        style={{
          background: 'var(--hr-card)',
          border: '1px solid var(--hr-border)',
          borderRadius: 12,
          overflowX: 'auto',
        }}
      >
        {/* Table header */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: COLUMNS,
            minWidth: MIN_WIDTH,
            padding: '12px 20px',
            borderBottom: '1px solid var(--hr-border)',
            background: 'var(--hr-hover)',
            height: HEADER_HEIGHT,
            alignItems: 'center',
          }}
        >
          {[
            t('payroll.table.employee'),
            t('payroll.table.hours'),
            t('payroll.table.basePay'),
            t('payroll.table.overtimePay'),
            t('payroll.table.deductions'),
            t('payroll.table.bonuses'),
            t('payroll.table.netPay'),
          ].map((col) => (
            <span
              key={col}
              style={{
                fontSize: 11,
                fontWeight: 600,
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                color: 'var(--hr-text-dim)',
              }}
            >
              {col}
            </span>
          ))}
        </div>

        {/* Virtualised rows */}
        <div
          ref={parentRef}
          style={{ overflowY: 'auto', overflowX: 'auto', maxHeight: 'calc(100vh - 420px)' }}
        >
          <div
            style={{
              height: virtualizer.getTotalSize(),
              width: '100%',
              position: 'relative',
              minWidth: MIN_WIDTH,
            }}
          >
            {virtualizer.getVirtualItems().map((virtualRow) => {
              const entry: PayrollEntryPayload = entries[virtualRow.index];
              const basePay = entry.basePay ?? 0;
              const overtimePay = entry.overtimePay ?? 0;
              const bonuses = entry.bonuses ?? 0;
              const deductions = entry.deductions ?? 0;
              const netPay = entry.netPay ?? 0;
              const grossPay = basePay + overtimePay + bonuses;
              const deductionsExceedGross = deductions > grossPay;
              return (
                <div
                  key={entry.id}
                  style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    width: '100%',
                    height: ROW_HEIGHT,
                    transform: `translateY(${virtualRow.start}px)`,
                    display: 'grid',
                    gridTemplateColumns: COLUMNS,
                    alignItems: 'center',
                    padding: '0 20px',
                    borderBottom: '1px solid var(--hr-border)',
                    background: deductionsExceedGross
                      ? 'color-mix(in srgb, var(--hr-error) 6%, transparent)'
                      : undefined,
                  }}
                >
                  {/* Employee ID */}
                  <span
                    style={{
                      fontSize: 13,
                      fontWeight: 600,
                      color: 'var(--hr-text)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 6,
                    }}
                  >
                    {t('payroll.table.employeeId', { id: entry.employeeId })}
                    {deductionsExceedGross && (
                      <span
                        style={{
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: 3,
                          padding: '2px 6px',
                          borderRadius: 4,
                          background: 'color-mix(in srgb, var(--hr-error) 15%, transparent)',
                          border: '1px solid color-mix(in srgb, var(--hr-error) 30%, transparent)',
                          color: 'var(--hr-error)',
                          fontSize: 10,
                          fontWeight: 700,
                          whiteSpace: 'nowrap',
                        }}
                      >
                        <AlertTriangle size={9} />
                        {t('payroll.table.deductionsExceedGross')}
                      </span>
                    )}
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--hr-text)' }}>
                    {entry.hoursWorked}h
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--hr-text)' }}>{fmt(basePay)}</span>
                  <span
                    style={{
                      fontSize: 13,
                      color:
                        overtimePay > 0 ? 'var(--hr-status-submitted)' : 'var(--hr-text-muted)',
                    }}
                  >
                    {overtimePay > 0 ? fmt(overtimePay) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color: deductions > 0 ? 'var(--hr-error)' : 'var(--hr-text-muted)',
                    }}
                  >
                    {deductions > 0 ? fmt(deductions) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color: bonuses > 0 ? 'var(--hr-status-approved)' : 'var(--hr-text-muted)',
                    }}
                  >
                    {bonuses > 0 ? fmt(bonuses) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      fontWeight: 700,
                      color: 'var(--hr-status-approved)',
                    }}
                  >
                    {fmt(netPay)}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      <Link
        to="/payroll"
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: 6,
          fontSize: 12,
          color: 'var(--hr-text-muted)',
          textDecoration: 'none',
        }}
      >
        <ArrowLeft size={12} />
        {t('payroll.backToRuns')}
      </Link>
    </div>
  );
};

const PayrollRunDetail = (): React.JSX.Element => {
  return (
    <>
      <HelmetContainer title="Payroll run · rtylr" />
      <ErrorBoundary>
        <PayrollRunDetailContent />
      </ErrorBoundary>
    </>
  );
};

export default PayrollRunDetail;
