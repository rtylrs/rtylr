import React, { useState } from 'react';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { CheckCircle, AlertTriangle } from 'lucide-react';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { useBalanceSheet } from './useBalanceSheet';
import { useErpInventoryValuation } from '../../shared/hooks/queries/useErpInventoryValuation';
import type { BalanceSheetLine } from '../../shared/schemas/balanceSheet';
import { cardStyle, ghostBtn, inputStyle } from '../../shared/lib/tokens';
import { formatDate, todayIso } from '../../shared/lib/date';
import { formatMoney } from '../../shared/lib/money';
import { useSettings } from '../../shared/lib/settings';

interface Row {
  key: string;
  label: string;
  code?: string;
  amount: number;
}

const Section: React.FC<{
  title: string;
  totalLabel: string;
  rows: Row[];
  total: number;
  emptyText: string;
}> = ({ title, totalLabel, rows, total, emptyText }) => {
  const currency = useSettings((s) => s.currency);
  return (
    <div style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column' }}>
      <h3 style={{ margin: '0 0 10px', fontSize: 14, fontWeight: 700, color: 'var(--hub-text)' }}>
        {title}
      </h3>
      <div style={{ flex: 1 }}>
        {rows.length === 0 ? (
          <p style={{ margin: '6px 0', fontSize: 13, color: 'var(--hub-text-dim)' }}>{emptyText}</p>
        ) : (
          rows.map((r) => (
            <div
              key={r.key}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                gap: 12,
                padding: '8px 0',
                borderBottom: '1px solid var(--hub-border)',
                fontSize: 13,
              }}
            >
              <span style={{ color: 'var(--hub-text-muted)' }}>
                {r.code && (
                  <span
                    style={{
                      marginInlineEnd: 8,
                      color: 'var(--hub-text-dim)',
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    {r.code}
                  </span>
                )}
                {r.label}
              </span>
              <span
                style={{
                  color: 'var(--hub-text)',
                  fontWeight: 600,
                  fontVariantNumeric: 'tabular-nums',
                }}
              >
                {formatMoney(r.amount, currency)}
              </span>
            </div>
          ))
        )}
      </div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          paddingTop: 12,
          fontSize: 14,
          fontWeight: 700,
          color: 'var(--hub-text)',
          fontVariantNumeric: 'tabular-nums',
        }}
      >
        <span>{totalLabel}</span>
        <span>{formatMoney(total, currency)}</span>
      </div>
    </div>
  );
};

const toRows = (lines: BalanceSheetLine[], type: BalanceSheetLine['type']): Row[] =>
  lines
    .filter((l) => l.type === type)
    .map((l) => ({ key: l.accountId, label: l.name, code: l.code, amount: l.amount }));

const BalanceSheet: React.FC = () => {
  const { t } = useTranslation('finance');
  const currency = useSettings((s) => s.currency);
  const [asOf, setAsOf] = useState(todayIso);
  const { data, isLoading, isError, isFetching, refetch } = useBalanceSheet(asOf);
  // ERP->FIN-02: inventory is shown only when ERP answers; it is valued
  // outside the ledger, so it sits with the other figures beside the statement.
  const { data: inventory } = useErpInventoryValuation();

  const header = (
    <PageHeader
      title={t('finance.balanceSheet.title')}
      subtitle={t('finance.balanceSheet.subtitle')}
      breadcrumbs={[{ label: t('financeNav.reports') }, { label: t('financeNav.balanceSheet') }]}
      actions={
        <label
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            fontSize: 12,
            color: 'var(--hub-text-muted)',
          }}
        >
          {t('finance.balanceSheet.asOfLabel')}
          <input
            type="date"
            value={asOf}
            max={todayIso()}
            onChange={(e) => e.target.value && setAsOf(e.target.value)}
            style={{ ...inputStyle, width: 'auto' }}
          />
        </label>
      }
    />
  );

  if (isLoading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        {header}
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      </div>
    );
  }
  if (isError || !data) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        {header}
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text)' }}>
            {t('financeErrors.generic')}
          </p>
          <button
            type="button"
            style={{ ...ghostBtn, alignSelf: 'flex-start' }}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      </div>
    );
  }

  const equityRows = toRows(data.lines, 'equity');
  if (data.netIncome !== 0) {
    equityRows.push({
      key: 'net-income',
      label: t('finance.balanceSheet.netIncome'),
      amount: data.netIncome,
    });
  }
  const difference = data.totalAssets - (data.totalLiabilities + data.totalEquity);
  const balanced = Math.abs(difference) < 0.005;
  const nothingPosted = data.lines.length === 0 && data.netIncome === 0;
  const dateLabel = formatDate(data.asOf || asOf, {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });

  const memo = [
    { key: 'bank', label: t('finance.balanceSheet.memo.bank'), amount: data.cashBalance },
    {
      key: 'receivable',
      label: t('finance.balanceSheet.memo.receivable'),
      amount: data.accountsReceivable,
    },
    { key: 'payable', label: t('finance.balanceSheet.memo.payable'), amount: data.accountsPayable },
    ...(inventory
      ? [
          {
            key: 'inventory',
            label: t('finance.balanceSheet.inventory'),
            amount: inventory.totalValue,
          },
        ]
      : []),
  ];

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeCommon.pageTitles.balanceSheet'))} />
      {header}

      {nothingPosted && (
        <div
          style={{
            ...cardStyle,
            padding: 16,
            fontSize: 13,
            color: 'var(--hub-text-muted)',
            lineHeight: 1.6,
          }}
        >
          {t('finance.balanceSheet.emptyNote', { date: dateLabel })}
        </div>
      )}

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
          gap: 16,
        }}
      >
        <Section
          title={t('finance.balanceSheet.assets')}
          totalLabel={t('finance.balanceSheet.totalAssets')}
          rows={toRows(data.lines, 'asset')}
          total={data.totalAssets}
          emptyText={t('finance.balanceSheet.noBalances')}
        />
        <Section
          title={t('finance.balanceSheet.liabilities')}
          totalLabel={t('finance.balanceSheet.totalLiabilities')}
          rows={toRows(data.lines, 'liability')}
          total={data.totalLiabilities}
          emptyText={t('finance.balanceSheet.noBalances')}
        />
        <Section
          title={t('finance.balanceSheet.equity')}
          totalLabel={t('finance.balanceSheet.totalEquity')}
          rows={equityRows}
          total={data.totalEquity}
          emptyText={t('finance.balanceSheet.noBalances')}
        />
      </div>

      <div
        role="status"
        style={{
          ...cardStyle,
          padding: '14px 20px',
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          fontSize: 13,
          color: 'var(--hub-text)',
        }}
      >
        {balanced ? (
          <CheckCircle size={16} style={{ color: 'var(--finance-success)', flexShrink: 0 }} />
        ) : (
          <AlertTriangle size={16} style={{ color: 'var(--finance-danger)', flexShrink: 0 }} />
        )}
        {balanced
          ? t('finance.balanceSheet.balanced', {
              assets: formatMoney(data.totalAssets, currency),
              liabilitiesAndEquity: formatMoney(data.totalLiabilities + data.totalEquity, currency),
            })
          : t('finance.balanceSheet.outOfBalance', {
              difference: formatMoney(difference, currency),
            })}
      </div>

      <div style={{ ...cardStyle, padding: 20 }}>
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
          {t('finance.balanceSheet.memo.title')}
        </h3>
        <p style={{ margin: '4px 0 10px', fontSize: 12, color: 'var(--hub-text-dim)' }}>
          {t('finance.balanceSheet.memo.description')}
        </p>
        {memo.map((m) => (
          <div
            key={m.key}
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '7px 0',
              borderTop: '1px solid var(--hub-border)',
              fontSize: 13,
            }}
          >
            <span style={{ color: 'var(--hub-text-muted)' }}>{m.label}</span>
            <span
              style={{
                fontWeight: 600,
                color: 'var(--hub-text)',
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              {formatMoney(m.amount, currency)}
            </span>
          </div>
        ))}
      </div>
    </motion.div>
  );
};

export default BalanceSheet;
