import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { AlertTriangle, RotateCcw } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import CashFlowChart from '../../shared/components/CashFlowChart';
import PeriodSelector from '../../shared/components/PeriodSelector';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useCashFlow } from './useCashFlow';
import { cardStyle, ghostBtn, ACCENT } from '../../shared/lib/tokens';

const CashFlow: React.FC = () => {
  const { t } = useTranslation();
  const [period, setPeriod] = useState<'month' | 'quarter' | 'year'>('month');
  const { data, isPending, isError, isFetching, refetch } = useCashFlow({ period });
  const cashFlowData = data ?? [];
  const loading = isPending || (!data && !isError);

  const net = cashFlowData.reduce((s: number, d) => s + (d.inflow - d.outflow), 0);
  const totalIn = cashFlowData.reduce((s: number, d) => s + d.inflow, 0);
  const totalOut = cashFlowData.reduce((s: number, d) => s + d.outflow, 0);

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={`${t('finance.cashFlow.title')} · rtylr`} />
      <PageHeader
        title={t('finance.cashFlow.title')}
        breadcrumbs={[
          { label: t('finance.cashFlow.reports') },
          { label: t('finance.cashFlow.title') },
        ]}
        actions={<PeriodSelector period={period} onPeriodChange={setPeriod} />}
      />

      {loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : isError ? (
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <AlertTriangle size={16} style={{ color: 'var(--finance-danger)' }} />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('errors.generic')}
            </span>
          </div>
          <button
            type="button"
            style={ghostBtn}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            <RotateCcw size={13} /> {t('common.refresh')}
          </button>
        </div>
      ) : (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
            <div style={{ ...cardStyle, padding: 20 }}>
              <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
                {t('finance.cashFlow.totalInflow')}
              </div>
              <div style={{ fontSize: 22, fontWeight: 700, color: ACCENT, marginTop: 6 }}>
                <MoneyDisplay amount={totalIn} />
              </div>
            </div>
            <div style={{ ...cardStyle, padding: 20 }}>
              <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
                {t('finance.cashFlow.totalOutflow')}
              </div>
              <div
                style={{
                  fontSize: 22,
                  fontWeight: 700,
                  color: 'var(--finance-danger)',
                  marginTop: 6,
                }}
              >
                <MoneyDisplay amount={totalOut} />
              </div>
            </div>
            <div style={{ ...cardStyle, padding: 20 }}>
              <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>
                {t('finance.cashFlow.netCashFlow')}
              </div>
              <div
                style={{
                  fontSize: 22,
                  fontWeight: 700,
                  color: net >= 0 ? ACCENT : 'var(--finance-danger)',
                  marginTop: 6,
                }}
              >
                <MoneyDisplay amount={net} signed />
              </div>
            </div>
          </div>

          <div style={{ ...cardStyle, padding: 24 }}>
            <CashFlowChart data={cashFlowData} height={240} />
          </div>
        </>
      )}
    </motion.div>
  );
};

export default CashFlow;
