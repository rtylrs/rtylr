import React from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { getRouteApi, Link, useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { ArrowLeft, FileText } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useGetMyPayrollRun } from '../../shared/hooks/queries/usePayroll';
import type { PayrollEntryPayload } from '../../shared/hooks/queries/usePayroll';
import { Card, SectionHeader, Skeleton, Breadcrumbs, Pill } from '../../shared/components/ui';
import { useHrSettings } from '../../shared/hooks/useHrExtras';
import { formatCurrency } from '../../shared/lib/hrNumberUtils';
import {
  PayrollConfigRequired,
  usePayrollConfigured,
} from '../../shared/components/PayrollConfigGuard';

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

// Route id, not path: `hrConfigRoute` is pathless with `id: 'hr-config'`, so
// every feature route is registered beneath it.
const routeApi = getRouteApi('/business/hr/hr-config/payroll/my-payslips/view');

const PaySlip = (): React.JSX.Element => {
  const { t } = useTranslation('hrCommon');
  const navigate = useNavigate();
  const { data: settings } = useHrSettings();
  const { runId } = routeApi.useSearch();
  const payrollConfigured = usePayrollConfigured();

  const { data: run, isLoading } = useGetMyPayrollRun(runId, { enabled: payrollConfigured });

  if (!payrollConfigured) return <PayrollConfigRequired />;

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <Skeleton height={32} width={200} />
        <Skeleton height={300} />
      </div>
    );
  }

  // `/hr/payroll/my-runs/{id}` returns exactly the caller's own entry, so there
  // is nothing to match on here, and nothing else to accidentally render.
  const entry: PayrollEntryPayload | undefined = run?.entries?.[0];

  if (!run || !entry) {
    return (
      <EmptyState
        icon={FileText}
        title={t('hrPayroll.paySlip.notFound')}
        description={t('hrPayroll.paySlip.notFoundDescription')}
        action={{
          label: t('hrPayroll.paySlip.backToSlips'),
          onClick: () => navigate({ to: '/hr/payroll/my-payslips' }),
        }}
      />
    );
  }

  const basePay = entry.basePay ?? 0;
  const overtimePay = entry.overtimePay ?? 0;
  const bonuses = entry.bonuses ?? 0;
  const deductions = entry.deductions ?? 0;
  const netPay = entry.netPay ?? 0;
  const gross = basePay + overtimePay + bonuses;

  return (
    <>
      <HelmetContainer title={pageTitle(t('hrCommon.pageTitles.paySlip'))} />
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20, maxWidth: 600 }}>
        <Breadcrumbs
          items={[
            { label: t('hrPayroll.paySlip.myPaySlips'), href: '/hr/payroll/my-payslips' },
            { label: `${run.periodStart}, ${run.periodEnd}` },
          ]}
        />

        <SectionHeader
          title={t('hrPayroll.paySlip.title')}
          subtitle={`${run.periodStart}, ${run.periodEnd}`}
        />

        <Card>
          <div style={{ marginBottom: 16 }}>
            <div style={{ fontSize: 15, fontWeight: 700, color: 'var(--hr-text)' }}>
              {t('hrPayroll.table.employeeId', { id: entry.employeeId })}
            </div>
            <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 2 }}>
              {t('hrPayroll.paySlip.period')}: {run.periodStart}, {run.periodEnd}
            </div>
            <Pill color="var(--hr-status-approved)" style={{ marginTop: 8 }}>
              {t(`hrPayroll.statuses.${run.status}`)}
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
              {t('hrPayroll.paySlip.earnings')}
            </div>
            <BreakdownRow
              label={t('hrPayroll.table.basePay')}
              value={formatCurrency(basePay, settings?.currency)}
            />
            <BreakdownRow
              label={t('hrPayroll.table.overtimePay')}
              value={formatCurrency(overtimePay, settings?.currency)}
              color={overtimePay > 0 ? 'var(--hr-status-submitted)' : undefined}
            />
            <BreakdownRow
              label={t('hrPayroll.table.bonuses')}
              value={formatCurrency(bonuses, settings?.currency)}
              color={bonuses > 0 ? 'var(--hr-status-approved)' : undefined}
            />
            <BreakdownRow
              label={t('hrPayroll.paySlip.grossPay')}
              value={formatCurrency(gross, settings?.currency)}
              bold
            />
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
              {t('hrPayroll.paySlip.deductionsSection')}
            </div>
            <BreakdownRow
              label={t('hrPayroll.table.deductions')}
              value={`-${formatCurrency(deductions, settings?.currency)}`}
              color="var(--hr-error)"
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
              {t('hrPayroll.table.netPay')}
            </span>
            <span
              style={{
                fontSize: 18,
                fontWeight: 800,
                color: 'var(--hr-status-approved)',
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {formatCurrency(netPay, settings?.currency)}
            </span>
          </div>
        </Card>

        <Link
          to="/hr/payroll/my-payslips"
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
          {t('hrPayroll.paySlip.backToSlips')}
        </Link>
      </div>
    </>
  );
};

export default PaySlip;
