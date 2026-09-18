import React from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { Link } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { useTranslation } from 'react-i18next';
import { FileText, DollarSign } from 'lucide-react';
import { EmptyState } from '@voxire/ui';
import { useListMyPayrollRuns } from '../../shared/hooks/queries/usePayroll';
import type { PayrollRunPayload } from '../../shared/hooks/queries/usePayroll';
import { SectionHeader, Card, Pill, Skeleton } from '../../shared/components/ui';
import { useHrSettings } from '../../shared/hooks/useHrExtras';
import { formatCurrency } from '../../shared/lib/hrNumberUtils';
import {
  PayrollConfigRequired,
  usePayrollConfigured,
} from '../../shared/components/PayrollConfigGuard';

const MyPaySlips = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: settings } = useHrSettings();
  const payrollConfigured = usePayrollConfigured();
  const { data, isLoading, refetch } = useListMyPayrollRuns(
    { limit: 50, offset: 0 },
    { enabled: payrollConfigured }
  );
  const runs = data?.runs ?? [];

  if (!payrollConfigured) return <PayrollConfigRequired />;

  return (
    <>
      <HelmetContainer title={pageTitle(t('nav.myPaySlips'))} />
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
            icon={FileText}
            title={t('payroll.paySlip.emptyTitle')}
            description={t('payroll.paySlip.emptyDescription')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {runs.map((run: PayrollRunPayload) => (
              <Link
                key={run.id}
                to="/payroll/my-payslips/view"
                // `PayrollRunPayload.id` is an `int64`, which the contract
                // generator emits as `string`. The route carries it as a
                // number; a run with no id yields NaN, which the search schema
                // resolves to `undefined` and renders as the empty state.
                search={{ runId: Number(run.id) }}
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

export default MyPaySlips;
