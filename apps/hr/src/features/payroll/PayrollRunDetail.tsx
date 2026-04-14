import React, { useRef, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { CheckCircle, Download, ArrowLeft } from 'lucide-react';
import {
  useGetPayrollRun,
  useApprovePayrollRun,
  useExportPayrollRun,
} from '../../shared/hooks/queries/usePayroll';
import type { PayrollRunEntry } from '../../shared/hooks/queries/usePayroll';
import ErrorBoundary from '../../shared/components/ErrorBoundary';
import { SectionHeader, Button, Skeleton, Pill, Breadcrumbs } from '../../shared/components/ui';

const ACCENT = '#8b5cf6';

const STATUS_COLORS: Record<string, string> = {
  Draft: 'var(--hr-text-muted)',
  Pending: 'var(--hr-status-submitted)',
  Approved: 'var(--hr-status-approved)',
  Paid: '#10b981',
};

function fmt(n: number): string {
  return n.toLocaleString('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
  });
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

const ROW_HEIGHT = 56;
const HEADER_HEIGHT = 44;
const COLUMNS = '2fr 1fr 1fr 1fr 1fr 1fr 1fr';
const MIN_WIDTH = 860;

const PayrollRunDetailContent = (): React.JSX.Element => {
  const { id } = useParams<{ id: string }>();
  const runId = id ? Number(id) : undefined;
  const { t } = useTranslation();
  const { data: run, isLoading } = useGetPayrollRun(runId);
  const approveRun = useApprovePayrollRun();
  const exportRun = useExportPayrollRun();
  const parentRef = useRef<HTMLDivElement>(null);

  const entries = run?.entries ?? [];

  const virtualizer = useVirtualizer({
    count: entries.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: 10,
  });

  const handleApprove = useCallback((): void => {
    if (runId) approveRun.mutate({ runId, approved: true });
  }, [runId, approveRun]);

  const handleExport = useCallback(
    (format: 'csv' | 'pdf'): void => {
      if (runId) exportRun.mutate({ runId, format });
    },
    [runId, exportRun]
  );

  if (isLoading || !run) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={400} />
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <Breadcrumbs
        items={[
          { label: t('payroll.title'), href: '/payroll' },
          { label: `${run.periodStart} — ${run.periodEnd}` },
        ]}
      />

      <SectionHeader
        title={`${run.periodStart} — ${run.periodEnd}`}
        subtitle={t('payroll.runDetail.subtitle', { count: entries.length })}
        right={
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <Pill color={STATUS_COLORS[run.status]}>
              {t(`payroll.statuses.${run.status.toLowerCase()}`)}
            </Pill>
            {run.status === 'Pending' && (
              <Button
                variant="primary"
                icon={<CheckCircle size={13} />}
                onClick={handleApprove}
                disabled={approveRun.isPending}
              >
                {t('payroll.runDetail.approve')}
              </Button>
            )}
            <Button
              variant="secondary"
              icon={<Download size={13} />}
              onClick={() => handleExport('csv')}
              disabled={exportRun.isPending}
            >
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
          { label: t('payroll.stats.totalBase'), value: fmt(run.totalBasePay), color: ACCENT },
          {
            label: t('payroll.stats.totalOvertime'),
            value: fmt(run.totalOvertimePay),
            color: '#f59e0b',
          },
          {
            label: t('payroll.stats.totalDeductions'),
            value: fmt(run.totalDeductions),
            color: '#ef4444',
          },
          { label: t('payroll.stats.totalNet'), value: fmt(run.totalNetPay), color: '#10b981' },
        ].map((stat) => (
          <div
            key={stat.label}
            style={{
              background: `${stat.color}12`,
              border: `1px solid ${stat.color}30`,
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
          overflow: 'hidden',
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
              const entry: PayrollRunEntry = entries[virtualRow.index];
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
                  }}
                >
                  {/* Employee name */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <span
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        height: 32,
                        width: 32,
                        borderRadius: '50%',
                        background: 'rgba(139,92,246,0.12)',
                        color: ACCENT,
                        fontSize: 11,
                        fontWeight: 700,
                        flexShrink: 0,
                      }}
                    >
                      {getInitials(entry.employeeName)}
                    </span>
                    <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hr-text)' }}>
                      {entry.employeeName}
                    </span>
                  </div>
                  <span style={{ fontSize: 13, color: 'var(--hr-text)' }}>
                    {entry.hoursWorked}h
                  </span>
                  <span style={{ fontSize: 13, color: 'var(--hr-text)' }}>
                    {fmt(entry.basePay)}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color: entry.overtimePay > 0 ? '#f59e0b' : 'var(--hr-text-muted)',
                    }}
                  >
                    {entry.overtimePay > 0 ? fmt(entry.overtimePay) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color: entry.deductions > 0 ? '#ef4444' : 'var(--hr-text-muted)',
                    }}
                  >
                    {entry.deductions > 0 ? fmt(entry.deductions) : '-'}
                  </span>
                  <span
                    style={{
                      fontSize: 13,
                      color: entry.bonuses > 0 ? '#10b981' : 'var(--hr-text-muted)',
                    }}
                  >
                    {entry.bonuses > 0 ? fmt(entry.bonuses) : '-'}
                  </span>
                  <span style={{ fontSize: 13, fontWeight: 700, color: '#10b981' }}>
                    {fmt(entry.netPay)}
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
    <ErrorBoundary>
      <PayrollRunDetailContent />
    </ErrorBoundary>
  );
};

export default PayrollRunDetail;
