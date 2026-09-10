import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import CashFlowChart from '../../shared/components/CashFlowChart';
import PeriodSelector from '../../shared/components/PeriodSelector';
import { useCashFlow } from './useCashFlow';
import { cardStyle, ACCENT } from '../../shared/lib/tokens';

const CashFlow: React.FC = () => {
  const { t } = useTranslation('finance');
  const [period, setPeriod] = useState<'month' | 'quarter' | 'year'>('month');
  const { data = [] } = useCashFlow({ period });

  const net = data.reduce((s: number, d) => s + (d.inflow - d.outflow), 0);
  const totalIn = data.reduce((s: number, d) => s + d.inflow, 0);
  const totalOut = data.reduce((s: number, d) => s + d.outflow, 0);

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
            style={{ fontSize: 22, fontWeight: 700, color: 'var(--finance-danger)', marginTop: 6 }}
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
        <CashFlowChart data={data} height={240} />
      </div>
    </motion.div>
  );
};

export default CashFlow;
