import React from 'react';
import { useTranslation } from 'react-i18next';
import { DollarSign } from 'lucide-react';
import { useHRConfigStore } from '../../../shared/stores/hrConfigStore';
import WidgetShell from './WidgetShell';
import { safeRatio, toFiniteNumber } from '@voxire/api';

interface PayrollCostData {
  currentMonth: number;
  previousMonth: number;
  currency: string;
}

interface PayrollCostWidgetProps {
  data: PayrollCostData | null;
  isLoading: boolean;
}

const PayrollCostWidget = ({ data, isLoading }: PayrollCostWidgetProps): React.JSX.Element => {
  const { t } = useTranslation();
  const config = useHRConfigStore((s) => s.config);

  if (!config?.payrollPeriod) {
    return (
      <WidgetShell
        title={t('dashboard.widgets.payrollCost')}
        icon={<DollarSign size={14} />}
        isLoading={false}
        isEmpty={true}
        emptyMessage={t('dashboard.widgets.payrollNotConfigured')}
      >
        <span />
      </WidgetShell>
    );
  }

  const currentMonth = toFiniteNumber(data?.currentMonth);
  const previousMonth = toFiniteNumber(data?.previousMonth);
  const changePercent =
    data && previousMonth > 0
      ? (safeRatio(currentMonth - previousMonth, previousMonth) * 100).toFixed(1)
      : null;

  return (
    <WidgetShell
      title={t('dashboard.widgets.payrollCost')}
      icon={<DollarSign size={14} />}
      isLoading={isLoading}
      isEmpty={!data}
      emptyMessage={t('dashboard.widgets.noData')}
    >
      {data && (
        <>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
            <span style={{ fontSize: 28, fontWeight: 700, color: 'var(--hr-text)' }}>
              {data.currency}
              {currentMonth.toLocaleString()}
            </span>
          </div>
          {changePercent !== null && (
            <div
              style={{
                fontSize: 11,
                fontWeight: 600,
                color: Number(changePercent) > 0 ? 'var(--hr-error)' : 'var(--hr-status-approved)',
              }}
            >
              {Number(changePercent) > 0 ? '+' : ''}
              {changePercent}% {t('dashboard.widgets.vsLastMonth')}
            </div>
          )}
        </>
      )}
    </WidgetShell>
  );
};

export default PayrollCostWidget;
