import React from 'react';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { FileText, DollarSign } from 'lucide-react';
import { useListPayrollRuns } from '../../shared/hooks/queries/usePayroll';
import type { PayrollRunPayload } from '../../shared/hooks/queries/usePayroll';
import { SectionHeader, Card, Pill, Skeleton, EmptyState } from '../../shared/components/ui';

function fmt(n: number): string {
  return n.toLocaleString('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
  });
}

const MyPaySlips = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading } = useListPayrollRuns({ status: 'exported', limit: 50, offset: 0 });
  const runs = data?.runs ?? [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <SectionHeader
        title={t('payroll.paySlip.myPaySlips')}
        subtitle={t('payroll.paySlip.myPaySlipsSubtitle')}
      />

      {isLoading ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} height={72} />
          ))}
        </div>
      ) : runs.length === 0 ? (
        <EmptyState
          icon={<FileText size={20} />}
          title={t('payroll.paySlip.emptyTitle')}
          description={t('payroll.paySlip.emptyDescription')}
        />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {runs.map((run: PayrollRunPayload) => (
            <Link
              key={run.id}
              to={`/payroll/my-payslips/view?runId=${run.id}` as any}
              style={{ textDecoration: 'none', color: 'inherit' }}
            >
              <Card
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 16,
                  cursor: 'pointer',
                }}
              >
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    height: 36,
                    width: 36,
                    borderRadius: 10,
                    background: 'color-mix(in srgb, var(--hr-status-approved) 12%, transparent)',
                    color: 'var(--hr-status-approved)',
                    flexShrink: 0,
                  }}
                >
                  <DollarSign size={16} />
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--hr-text)' }}>
                      {run.periodStart}, {run.periodEnd}
                    </span>
                    <Pill color="var(--hr-status-approved)">
                      {t(`payroll.statuses.${run.status}`)}
                    </Pill>
                  </div>
                  <div style={{ fontSize: 11, color: 'var(--hr-text-muted)', marginTop: 3 }}>
                    {t('payroll.netTotal')}: {fmt(run.totalNet)}
                  </div>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
};

export default MyPaySlips;
