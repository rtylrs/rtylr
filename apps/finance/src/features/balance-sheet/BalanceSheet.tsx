import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import AccountTypeBadge, { type AccountType } from '../../shared/components/AccountTypeBadge';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useBalanceSheet } from './useBalanceSheet';
import { cardStyle, ACCENT } from '../../shared/lib/tokens';

const Section: React.FC<{
  title: string;
  rows: { label: string; amount: number; type: string }[];
}> = ({ title, rows }) => {
  const total = rows.reduce((s, r) => s + r.amount, 0);
  return (
    <div style={{ ...cardStyle, padding: 20 }}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 14,
        }}
      >
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, color: 'var(--hub-text)' }}>
          {title}
        </h3>
        <AccountTypeBadge type={(rows[0]?.type ?? 'asset') as AccountType} />
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {rows.map((r) => (
          <div
            key={r.label}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '8px 0',
              borderBottom: '1px solid var(--hub-border)',
              fontSize: 13,
              color: 'var(--hub-text-muted)',
            }}
          >
            <span>{r.label}</span>
            <span style={{ color: 'var(--hub-text)', fontWeight: 600 }}>
              <MoneyDisplay amount={r.amount} />
            </span>
          </div>
        ))}
      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          paddingTop: 12,
          fontSize: 14,
          fontWeight: 700,
          color: ACCENT,
        }}
      >
        <span>Total {title}</span>
        <MoneyDisplay amount={total} />
      </div>
    </div>
  );
};

const BalanceSheet: React.FC = () => {
  const { data, isLoading } = useBalanceSheet();

  if (isLoading || !data) {
    return (
      <div>
        <PageHeader
          title="Balance Sheet"
          breadcrumbs={[{ label: 'Reports' }, { label: 'Balance Sheet' }]}
        />
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      </div>
    );
  }

  const totalAssets = data.assets.reduce((s: number, r) => s + r.amount, 0);
  const totalLiab = data.liabilities.reduce((s: number, r) => s + r.amount, 0);
  const totalEquity = data.equity.reduce((s: number, r) => s + r.amount, 0);

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title="Balance sheet · rtylr" />
      <PageHeader
        title="Balance Sheet"
        subtitle="As of today"
        breadcrumbs={[{ label: 'Reports' }, { label: 'Balance Sheet' }]}
      />

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
          gap: 16,
        }}
      >
        <Section title="Assets" rows={data.assets} />
        <Section title="Liabilities" rows={data.liabilities} />
        <Section title="Equity" rows={data.equity} />
      </div>

      <div
        style={{
          ...cardStyle,
          padding: 20,
          display: 'flex',
          justifyContent: 'space-around',
          alignItems: 'center',
        }}
      >
        <Stat label="Total Assets" value={totalAssets} />
        <Stat label="Total Liabilities" value={totalLiab} color="var(--finance-danger)" />
        <Stat label="Total Equity" value={totalEquity} color="var(--finance-purple)" />
        <Stat label="Check" value={totalAssets - (totalLiab + totalEquity)} color={ACCENT} />
      </div>
    </motion.div>
  );
};

const Stat: React.FC<{ label: string; value: number; color?: string }> = ({
  label,
  value,
  color = 'var(--hub-text)',
}) => (
  <div style={{ textAlign: 'center' }}>
    <div
      style={{
        fontSize: 11,
        color: 'var(--hub-text-dim)',
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
      }}
    >
      {label}
    </div>
    <div style={{ fontSize: 18, fontWeight: 700, color, marginTop: 4 }}>
      <MoneyDisplay amount={value} />
    </div>
  </div>
);

export default BalanceSheet;
