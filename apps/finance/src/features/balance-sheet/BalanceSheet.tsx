import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import PageHeader from '../../shared/components/PageHeader';
import MoneyDisplay from '../../shared/components/MoneyDisplay';
import AccountTypeBadge, { type AccountType } from '../../shared/components/AccountTypeBadge';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useBalanceSheet } from './useBalanceSheet';
import { useErpInventoryValuation } from '../../shared/hooks/queries/useErpInventoryValuation';
import { cardStyle, ACCENT } from '../../shared/lib/tokens';
import { todayIso } from '../../shared/lib/date';

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
  const { t } = useTranslation('finance');
  const [asOfDate, setAsOfDate] = useState(todayIso());
  const { data, isLoading } = useBalanceSheet(asOfDate);
  // ERP->FIN-02: the Inventory line exists only when ERP answers. A tenant
  // without ERP gets no section at all, never a zero: a balance sheet that
  // silently reports 0 inventory misstates the business (#845).
  const { data: inventory } = useErpInventoryValuation();

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

  const erpInventory = inventory?.totalValue ?? 0;
  const totalAssets = (data.totalAssets ?? 0) + erpInventory;
  const totalLiab = data.totalLiabilities ?? 0;
  const totalEquity = data.totalEquity ?? 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('common.pageTitles.balanceSheet'))} />
      <PageHeader
        title="Balance Sheet"
        subtitle={`As of ${asOfDate}`}
        breadcrumbs={[{ label: 'Reports' }, { label: 'Balance Sheet' }]}
        actions={
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <label style={{ fontSize: 12, color: 'var(--hub-text-muted)', fontWeight: 500 }}>
              As of Date:
            </label>
            <input
              type="date"
              value={asOfDate}
              max={todayIso()}
              onChange={(e) => setAsOfDate(e.target.value)}
              style={{
                padding: '5px 10px',
                borderRadius: 8,
                border: '1px solid var(--hub-border)',
                background: 'var(--hub-surface)',
                color: 'var(--hub-text)',
                fontSize: 12,
              }}
            />
          </div>
        }
      />

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
          gap: 16,
        }}
      >
        <Section
          title="Assets"
          rows={
            inventory
              ? [
                  ...data.assets,
                  {
                    label: t('finance.balanceSheet.inventory'),
                    amount: inventory.totalValue,
                    type: 'asset',
                  },
                ]
              : data.assets
          }
        />
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
