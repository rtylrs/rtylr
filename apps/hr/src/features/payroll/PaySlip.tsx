import React from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ArrowLeft, FileText } from 'lucide-react';
import { useGetPayrollRun } from '../../shared/hooks/queries/usePayroll';
import type { PayrollRunEntry } from '../../shared/hooks/queries/usePayroll';
import {
  Card,
  SectionHeader,
  Skeleton,
  EmptyState,
  Breadcrumbs,
  Pill,
} from '../../shared/components/ui';

function fmt(n: number): string {
  return n.toLocaleString('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
  });
}

interface BreakdownRowProps {
  label: string;
  value: string;
  color?: string;
  bold?: boolean;
}

const BreakdownRow = ({ label, value, color, bold }: BreakdownRowProps): React.JSX.Element => (
  <div
    style={{
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      padding: '10px 0',
      borderBottom: '1px solid var(--hr-border)',
    }}
  >
    <span
      style={{
        fontSize: 13,
        fontWeight: bold ? 700 : 400,
        color: 'var(--hr-text)',
      }}
    >
      {label}
    </span>
    <span
      style={{
        fontSize: 13,
        fontWeight: bold ? 700 : 600,
        color: color ?? 'var(--hr-text)',
        fontVariantNumeric: 'tabular-nums',
      }}
    >
      {value}
    </span>
  </div>
);

const PaySlip = (): React.JSX.Element => {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const runId = searchParams.get('runId') ? Number(searchParams.get('runId')) : undefined;
  const employeeId = searchParams.get('employeeId')
    ? Number(searchParams.get('employeeId'))
    : undefined;

  const { data: run, isLoading } = useGetPayrollRun(runId);

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={300} />
      </div>
    );
  }

  const entry: PayrollRunEntry | undefined = run?.entries?.find(
    (e: PayrollRunEntry) => e.employeeId === employeeId
  );

  if (!run || !entry) {
    return (
      <EmptyState
        icon={<FileText size={20} />}
        title={t('payroll.paySlip.notFound')}
        description={t('payroll.paySlip.notFoundDescription')}
        action={
          <Link
            to="/payroll/my-payslips"
            style={{ color: 'var(--hr-accent)', textDecoration: 'none', fontSize: 12 }}
          >
            {t('payroll.paySlip.backToSlips')}
          </Link>
        }
      />
    );
  }

  const gross = entry.basePay + entry.overtimePay + entry.bonuses;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 600 }}>
      <Breadcrumbs
        items={[
          { label: t('payroll.paySlip.myPaySlips'), href: '/payroll/my-payslips' },
          { label: `${run.periodStart} — ${run.periodEnd}` },
        ]}
      />

      <SectionHeader
        title={t('payroll.paySlip.title')}
        subtitle={`${run.periodStart} — ${run.periodEnd}`}
      />

      <Card>
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 15, fontWeight: 700, color: 'var(--hr-text)' }}>
            {entry.employeeName}
          </div>
          <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 2 }}>
            {t('payroll.paySlip.period')}: {run.periodStart} — {run.periodEnd}
          </div>
          <Pill color="#10b981" style={{ marginTop: 8 }}>
            {t(`payroll.statuses.${run.status.toLowerCase()}`)}
          </Pill>
        </div>

        {/* Earnings section */}
        <div style={{ marginBottom: 16 }}>
          <div
            style={{
              fontSize: 11,
              fontWeight: 600,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: 'var(--hr-text-muted)',
              marginBottom: 8,
            }}
          >
            {t('payroll.paySlip.earnings')}
          </div>
          <BreakdownRow label={t('payroll.table.basePay')} value={fmt(entry.basePay)} />
          <BreakdownRow
            label={t('payroll.table.overtimePay')}
            value={fmt(entry.overtimePay)}
            color={entry.overtimePay > 0 ? '#f59e0b' : undefined}
          />
          <BreakdownRow
            label={t('payroll.table.bonuses')}
            value={fmt(entry.bonuses)}
            color={entry.bonuses > 0 ? '#10b981' : undefined}
          />
          <BreakdownRow label={t('payroll.paySlip.grossPay')} value={fmt(gross)} bold />
        </div>

        {/* Deductions section */}
        <div style={{ marginBottom: 16 }}>
          <div
            style={{
              fontSize: 11,
              fontWeight: 600,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              color: 'var(--hr-text-muted)',
              marginBottom: 8,
            }}
          >
            {t('payroll.paySlip.deductionsSection')}
          </div>
          <BreakdownRow
            label={t('payroll.table.deductions')}
            value={`-${fmt(entry.deductions)}`}
            color="#ef4444"
          />
        </div>

        {/* Net pay */}
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            padding: '14px 0',
            borderTop: '2px solid var(--hr-border)',
          }}
        >
          <span style={{ fontSize: 15, fontWeight: 700, color: 'var(--hr-text)' }}>
            {t('payroll.table.netPay')}
          </span>
          <span
            style={{
              fontSize: 18,
              fontWeight: 800,
              color: '#10b981',
              fontVariantNumeric: 'tabular-nums',
            }}
          >
            {fmt(entry.netPay)}
          </span>
        </div>
      </Card>

      <Link
        to="/payroll/my-payslips"
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
        {t('payroll.paySlip.backToSlips')}
      </Link>
    </div>
  );
};

export default PaySlip;
