import React, { useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import CashFlowChart from '../../shared/components/CashFlowChart';
import PeriodSelector from '../../shared/components/PeriodSelector';
import { useCashFlow } from './useCashFlow';
import { cardStyle, ACCENT } from '../../shared/lib/tokens';

const CashFlow: React.FC = () => {
  const { data = [] } = useCashFlow();
  const [period, setPeriod] = useState<'month' | 'quarter' | 'year'>('month');
  const [date, setDate] = useState(new Date());

  const net = data.reduce((s: number, d) => s + (d.inflow - d.outflow), 0);
  const totalIn = data.reduce((s: number, d) => s + d.inflow, 0);
  const totalOut = data.reduce((s: number, d) => s + d.outflow, 0);

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="Cash flow · rtylr" />
      <PageHeader
        title="Cash Flow"
        breadcrumbs={[{ label: 'Reports' }, { label: 'Cash Flow' }]}
        actions={
          <PeriodSelector
            period={period}
            onPeriodChange={setPeriod}
            date={date}
            onDateChange={setDate}
          />
        }
      />

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>Total inflow</div>
          <div style={{ fontSize: 22, fontWeight: 700, color: ACCENT, marginTop: 6 }}>
            <MoneyDisplay amount={totalIn} />
          </div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>Total outflow</div>
          <div
            style={{ fontSize: 22, fontWeight: 700, color: 'var(--finance-danger)', marginTop: 6 }}
          >
            <MoneyDisplay amount={totalOut} />
          </div>
        </div>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>Net cash flow</div>
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
