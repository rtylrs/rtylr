import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { TrendingUp, TrendingDown, Download, Lock } from 'lucide-react';
import { toast } from 'sonner';
import { usePLByPeriod } from '../../shared/hooks/useTax';
import PaymentReminderPanel from '../../shared/components/PaymentReminderPanel';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import { exportProfitLossCsv } from '../../shared/lib/exports';

import { useSettings } from '../../shared/lib/settings';
import { formatMoney, type Currency } from '../../shared/lib/money';

type Period = 'This Month' | 'Last Month' | 'This Quarter' | 'This Year';

interface PeriodData {
  revenue: number;
  cogs: number;
  operatingExpenses: number;
}

const defaultPeriodData: Record<Period, PeriodData> = {
  'This Month': { revenue: 0, cogs: 0, operatingExpenses: 0 },
  'Last Month': { revenue: 0, cogs: 0, operatingExpenses: 0 },
  'This Quarter': { revenue: 0, cogs: 0, operatingExpenses: 0 },
  'This Year': { revenue: 0, cogs: 0, operatingExpenses: 0 },
};

const fmt = (n: number, currency = 'USD') => formatMoney(n, currency as Currency, 0);

const pct = (n: number, of: number) => (of === 0 ? 'N/A' : `${((n / of) * 100).toFixed(1)}%`);

const ACCENT = 'var(--finance-accent)';

const cardStyle: React.CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
  borderRadius: 12,
  overflow: 'hidden',
};

interface RowProps {
  label: string;
  value: number;
  indent?: boolean;
  bold?: boolean;
  positive?: boolean;
  negative?: boolean;
  separator?: boolean;
  sub?: string;
  large?: boolean;
  currency?: string;
}

const PLRow: React.FC<RowProps> = ({
  label,
  value,
  indent,
  bold,
  positive,
  negative,
  separator,
  sub,
  large,
  currency = 'USD',
}) => {
  const color = positive
    ? 'var(--finance-accent)'
    : negative
      ? 'var(--finance-danger)'
      : 'var(--hub-text)';
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: large ? '12px 0' : '8px 0',
        borderTop: separator ? '1px solid var(--hub-border)' : 'none',
        paddingInlineStart: indent ? 16 : 0,
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'column' }}>
        <span
          style={{
            fontSize: large ? 15 : 13,
            fontWeight: bold ? 700 : 400,
            color: bold ? 'var(--hub-text)' : 'var(--hub-text-muted)',
          }}
        >
          {label}
        </span>
        {sub && (
          <span
            style={{
              fontSize: 11,
              color: 'var(--hub-text-muted)',
              paddingInlineStart: indent ? 16 : 0,
            }}
          >
            {sub}
          </span>
        )}
      </div>
      <span
        style={{
          fontSize: large ? 16 : 13,
          fontWeight: bold ? 700 : 500,
          color,
          fontVariantNumeric: 'tabular-nums',
        }}
      >
        {fmt(value, currency)}
      </span>
    </div>
  );
};

const PERIOD_PARAMS: Record<Period, string> = {
  'This Month': 'month',
  'Last Month': 'last_month',
  'This Quarter': 'quarter',
  'This Year': 'year',
};

const ProfitLoss: React.FC = () => {
  const { t } = useTranslation();
  const [period, setPeriod] = useState<Period>('This Month');
  const [showClose, setShowClose] = useState(false);
  const currency = useSettings((s) => s.currency);
  const { data: plData } = usePLByPeriod(PERIOD_PARAMS[period]);

  const data: PeriodData = plData
    ? {
        revenue: plData.revenue ?? 0,
        cogs: plData.cogs ?? 0,
        operatingExpenses: plData.operatingExpenses ?? 0,
      }
    : defaultPeriodData[period];

  const grossProfit = data.revenue - data.cogs;
  const grossMargin = pct(grossProfit, data.revenue);
  const netProfit = grossProfit - data.operatingExpenses;
  const netMargin = pct(netProfit, data.revenue);

  const periods: Period[] = ['This Month', 'Last Month', 'This Quarter', 'This Year'];

  // Summary cards
  const summaryItems = [
    {
      id: 'revenue',

      label: t('finance.profitLoss.revenue'),
      value: data.revenue,
      color: 'var(--finance-accent)',
      Icon: TrendingUp,
    },
    {
      id: 'grossProfit',

      label: t('finance.profitLoss.grossProfit'),
      value: grossProfit,
      color: 'var(--finance-info)',
      Icon: TrendingUp,
    },
    {
      id: 'netProfit',

      label: t('finance.profitLoss.netProfit'),
      value: netProfit,
      color: netProfit >= 0 ? 'var(--finance-accent)' : 'var(--finance-danger)',
      Icon: netProfit >= 0 ? TrendingUp : TrendingDown,
    },
    {
      id: 'netMargin',

      label: t('finance.profitLoss.netMargin'),
      value: parseFloat(netMargin),
      color: 'var(--finance-warning)',
      Icon: TrendingUp,
      isMeta: true,
      display: netMargin,
    },
  ];

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('nav.pl'))} />
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hub-text)' }}>
            {t('finance.profitLoss.title')}
          </h2>
          <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
            Financial statement for {period.toLowerCase()}
          </p>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            type="button"
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '7px 14px',
              borderRadius: 8,
              border: '1px solid var(--hub-border)',
              background: 'var(--hub-hover)',
              color: 'var(--hub-text-muted)',
              fontSize: 13,
              cursor: 'pointer',
            }}
            onClick={() =>
              exportProfitLossCsv({
                period,
                revenue: data.revenue,
                cogs: data.cogs,
                grossProfit,
                operatingExpenses: data.operatingExpenses,
                netProfit,
              })
            }
            aria-label={t('finance.profitLoss.exportAria')}
          >
            <Download size={14} />
            {t('finance.profitLoss.export')}
          </button>
          <button
            type="button"
            onClick={() => setShowClose(true)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '7px 14px',
              borderRadius: 8,
              border: '1px solid color-mix(in srgb, var(--finance-warning) 30%, transparent)',
              background: 'color-mix(in srgb, var(--finance-warning) 8%, transparent)',
              color: 'var(--finance-warning)',
              fontSize: 13,
              cursor: 'pointer',
              fontWeight: 600,
            }}
          >
            <Lock size={13} />
            {t('finance.profitLoss.closeFiscalYear')}
          </button>
        </div>
      </div>

      {/* Period selector */}
      <div
        style={{
          display: 'flex',
          gap: 4,
          padding: 4,
          background: 'var(--hub-surface)',
          borderRadius: 10,
          border: '1px solid var(--hub-border)',
          width: 'fit-content',
        }}
      >
        {periods.map((p) => (
          <button
            type="button"
            key={p}
            onClick={() => setPeriod(p)}
            style={{
              padding: '5px 16px',
              borderRadius: 7,
              border: 'none',
              fontSize: 12,
              fontWeight: 500,
              cursor: 'pointer',
              background: period === p ? ACCENT : 'transparent',
              color: period === p ? 'var(--finance-white)' : 'var(--hub-text-muted)',
              transition: 'all 0.15s',
            }}
          >
            {p}
          </button>
        ))}
      </div>

      {/* Summary cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
        {summaryItems.map(({ id, label, value, color, Icon, isMeta, display }) => (
          <div
            key={id}
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: 12,
              padding: 18,
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 10,
              }}
            >
              <span
                style={{
                  fontSize: 11,
                  color: 'var(--hub-text-muted)',
                  fontWeight: 600,
                  textTransform: 'uppercase',
                  letterSpacing: '0.07em',
                }}
              >
                {label}
              </span>
              <Icon size={14} style={{ color }} />
            </div>
            <div
              style={{ fontSize: 22, fontWeight: 700, color, fontVariantNumeric: 'tabular-nums' }}
            >
              {isMeta ? display : fmt(value, currency)}
            </div>
          </div>
        ))}
      </div>

      {/* P&L Statement */}
      <div style={cardStyle}>
        {/* Header row */}
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 24px',
            borderBottom: '1px solid var(--hub-border)',
            background: 'var(--hub-surface)',
          }}
        >
          <span
            style={{
              fontSize: 13,
              fontWeight: 700,
              color: 'var(--hub-text)',
              textTransform: 'uppercase',
              letterSpacing: '0.06em',
            }}
          >
            {t('finance.profitLoss.statement')}
          </span>
          <span style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{period}</span>
        </div>

        {/* Revenue section */}
        <div style={{ padding: '8px 0' }}>
          <div
            style={{
              padding: '6px 24px',
              fontSize: 10,
              fontWeight: 700,
              textTransform: 'uppercase',
              letterSpacing: '0.1em',
              color: 'var(--hub-text-dim)',
            }}
          >
            {t('finance.profitLoss.revenue')}
          </div>
          <PLRow
            label={t('finance.profitLoss.grossRevenue')}
            value={data.revenue}
            positive
            currency={currency}
          />
        </div>

        {/* COGS */}
        <div style={{ borderTop: '1px solid var(--hub-border)', padding: '8px 0' }}>
          <div
            style={{
              padding: '6px 24px',
              fontSize: 10,
              fontWeight: 700,
              textTransform: 'uppercase',
              letterSpacing: '0.1em',
              color: 'var(--hub-text-dim)',
            }}
          >
            {t('finance.profitLoss.cogs')}
          </div>
          <PLRow
            label={t('finance.profitLoss.directCosts')}
            value={data.cogs}
            indent
            negative
            currency={currency}
          />
        </div>

        {/* Gross Profit */}
        <PLRow
          label="= Gross Profit"
          value={grossProfit}
          bold
          separator
          positive={grossProfit >= 0}
          negative={grossProfit < 0}
          sub={`Gross margin: ${grossMargin}`}
          large
          currency={currency}
        />

        {/* Operating expenses */}
        <div style={{ borderTop: '1px solid var(--hub-border)', padding: '8px 0' }}>
          <div
            style={{
              padding: '6px 24px',
              fontSize: 10,
              fontWeight: 700,
              textTransform: 'uppercase',
              letterSpacing: '0.1em',
              color: 'var(--hub-text-dim)',
            }}
          >
            {t('finance.profitLoss.operatingExpenses')}
          </div>
          <PLRow
            label={t('finance.profitLoss.totalOperatingExpenses')}
            value={data.operatingExpenses}
            indent
            negative
            currency={currency}
          />
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              padding: '6px 24px',
              borderTop: '1px solid var(--hub-border)',
              marginTop: 4,
            }}
          >
            <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text-muted)' }}>
              {t('finance.profitLoss.totalOperating')}
            </span>
            <span
              style={{
                fontSize: 12,
                fontWeight: 600,
                color: 'var(--finance-danger)',
                fontVariantNumeric: 'tabular-nums',
              }}
            >
              -{fmt(data.operatingExpenses, currency)}
            </span>
          </div>
        </div>

        {/* Net Profit */}
        <PLRow
          label="= Net Profit"
          value={netProfit}
          bold
          separator
          positive={netProfit >= 0}
          negative={netProfit < 0}
          sub={`Net margin: ${netMargin}`}
          large
          currency={currency}
        />
      </div>

      {/* By-category + reminders */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
        <div style={{ ...cardStyle, padding: 20 }}>
          <div
            style={{
              fontSize: 11,
              color: 'var(--hub-text-dim)',
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              marginBottom: 12,
            }}
          >
            {t('finance.profitLoss.expensesByCategory')}
          </div>
          <p style={{ fontSize: 13, color: 'var(--hub-text-muted)', margin: 0 }}>
            {t('finance.profitLoss.categoryUnavailable')}
          </p>
        </div>
        <PaymentReminderPanel />
      </div>

      <ConfirmDialog
        open={showClose}
        onClose={() => setShowClose(false)}
        onConfirm={() => toast.info(t('finance.profitLoss.lockingComingSoon'))}
        title={t('finance.profitLoss.closeYear')}
        message={t('finance.profitLoss.closeYearConfirm')}
        confirmLabel={t('common.closeYear')}
        destructive
      />
    </motion.div>
  );
};

export default ProfitLoss;
