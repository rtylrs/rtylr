import React, { useRef, useCallback } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { useUserStore } from '../../../../shared/stores/userStore';
import { getRouteApi, Link, useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { BookUp2, CheckCircle, Download, ArrowLeft, AlertTriangle, DollarSign } from 'lucide-react';
import {
  useGetPayrollRun,
  useApprovePayrollRun,
  useExportPayrollRun,
} from '../../shared/hooks/queries/usePayroll';
import type { PayrollEntryPayload } from '../../shared/hooks/queries/usePayroll';
import { ErrorBoundary, EmptyState } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import { SectionHeader, Button, Skeleton, Pill, Breadcrumbs } from '../../shared/components/ui';
import { canHrPermission } from '../../shared/lib/hrPermissions';
import { useHrSettings } from '../../shared/hooks/useHrExtras';
import { formatCurrency } from '../../shared/lib/hrNumberUtils';
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

const ROW_HEIGHT = 56;
const HEADER_HEIGHT = 44;
const COLUMNS = '2fr 1fr 1fr 1fr 1fr 1fr 1fr';
const MIN_WIDTH = 860;

// Route id, not path: `hrConfigRoute` is pathless with `id: 'hr-config'`, so
// every feature route is registered beneath it.
const routeApi = getRouteApi('/business/hr/hr-config/payroll/$id');

const PayrollRunDetailContent = (): React.JSX.Element => {
  const { id: runId } = routeApi.useParams();
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();
  const { data: settings } = useHrSettings();
  const payrollConfigured = usePayrollConfigured();
  const canRunPayroll = canHrPermission('hr:payroll:execute');
  const { data: run, isLoading } = useGetPayrollRun(runId, { enabled: payrollConfigured });
  const approveRun = useApprovePayrollRun(payrollConfigured);
  // HR->FIN-01: posts the approved run to the finance journal server-side.
  const exportRun = useExportPayrollRun(payrollConfigured);
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

  const handlePostToJournal = useCallback((): void => {
    if (runId) exportRun.mutate(runId);
  }, [runId, exportRun]);

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

  if (!payrollConfigured) return <PayrollConfigRequired />;

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
        icon={DollarSign}
        title={t('hrPayroll.runDetail.notFound')}
        description={t('hrPayroll.runDetail.notFoundDescription')}
        action={{
          label: t('hrPayroll.backToRuns'),
          onClick: () => navigate({ to: '/hr/payroll' }),
        }}
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
          { label: t('hrPayroll.title'), href: '/hr/payroll' },
          { label: `${run.periodStart}, ${run.periodEnd}` },
        ]}
      />

      <SectionHeader
        title={`${run.periodStart}, ${run.periodEnd}`}
        subtitle={t('hrPayroll.runDetail.subtitle', { count: entries.length })}
        right={
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <Pill color={STATUS_COLORS[run.status ?? '']}>
              {t(`hrPayroll.statuses.${run.status}`)}
            </Pill>
            {canRunPayroll && run.status === 'in_review' && (
              <Button
                variant="primary"
                icon={<CheckCircle size={13} />}
                onClick={handleApprove}
                disabled={approveRun.isPending}
              >
                {t('hrPayroll.runDetail.approve')}
              </Button>
            )}
            {canRunPayroll && run.status === 'approved' && (
              <Button
                variant="primary"
                icon={<BookUp2 size={13} />}
                onClick={handlePostToJournal}
                disabled={exportRun.isPending}
              >
                {t('hrPayroll.runDetail.postToJournal')}
              </Button>
            )}
            {canRunPayroll && (
              <Button variant="secondary" icon={<Download size={13} />} onClick={handleExport}>
                {t('hrPayroll.runDetail.exportCsv')}
              </Button>
            )}
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
            label: t('hrPayroll.stats.totalGross'),
            value: formatCurrency(run.totalGross ?? 0, settings?.currency),
            color: 'var(--accent)',
          },
          {
            label: t('hrPayroll.stats.totalDeductions'),
            value: formatCurrency(totalDeductions, settings?.currency),
            color: 'var(--hr-error)',
          },
          {
            label: t('hrPayroll.stats.totalNet'),
            value: formatCurrency(run.totalNet ?? 0, settings?.currency),
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
            t('hrPayroll.table.employee'),
            t('hrPayroll.table.hours'),
            t('hrPayroll.table.basePay'),
            t('hrPayroll.table.overtimePay'),
            t('hrPayroll.table.deductions'),
            t('hrPayroll.table.bonuses'),
            t('hrPayroll.table.netPay'),
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
                    {t('hrPayroll.table.employeeId', { id: entry.employeeId })}
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
                        {t('hrPayroll.table.deductionsExceedGross')}
                      </span>
                    )}
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--hr-text)' }}>
                    {entry.hoursWorked}h
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--hr-text)' }}>
                    {formatCurrency(basePay, settings?.currency)}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color:
                        overtimePay > 0 ? 'var(--hr-status-submitted)' : 'var(--hr-text-muted)',
                    }}
                  >
                    {overtimePay > 0 ? formatCurrency(overtimePay, settings?.currency) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color: deductions > 0 ? 'var(--hr-error)' : 'var(--hr-text-muted)',
                    }}
                  >
                    {deductions > 0 ? formatCurrency(deductions, settings?.currency) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color: bonuses > 0 ? 'var(--hr-status-approved)' : 'var(--hr-text-muted)',
                    }}
                  >
                    {bonuses > 0 ? formatCurrency(bonuses, settings?.currency) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      fontWeight: 700,
                      color: 'var(--hr-status-approved)',
                    }}
                  >
                    {formatCurrency(netPay, settings?.currency)}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      <Link
        to="/hr/payroll"
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
        {t('hrPayroll.backToRuns')}
      </Link>
    </div>
  );
};

const PayrollRunDetail = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  return (
    <>
      <HelmetContainer title={pageTitle(t('hrCommon.pageTitles.payrollRun'))} />
      <ErrorBoundary>
        <PayrollRunDetailContent />
      </ErrorBoundary>
    </>
  );
};

export default PayrollRunDetail;
