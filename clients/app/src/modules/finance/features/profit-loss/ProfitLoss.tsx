import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { Link } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Clock, Download, TrendingDown } from 'lucide-react';
import { useUserStore } from '../../../../shared/stores/userStore';
import { canFinancePermission } from '../../shared/lib/financePermissions';
import { usePLByPeriod } from '../../shared/hooks/useTax';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { exportProfitLossCsv } from '../../shared/lib/exports';
import { cardStyle, ghostBtn, ACCENT } from '../../shared/lib/tokens';
import { useSettings } from '../../shared/lib/settings';
import { formatMoney, type Currency } from '../../shared/lib/money';

type Period = 'month' | 'last_month' | 'quarter' | 'year';
const PERIODS: Period[] = ['month', 'last_month', 'quarter', 'year'];

// A margin needs revenue to be a margin; without any it is shown as a dash.
const pct = (n: number, of: number): string => (of === 0 ? '—' : `${((n / of) * 100).toFixed(1)}%`);

const sectionLabel: React.CSSProperties = {
  padding: '6px 24px',
  fontSize: 10,
  fontWeight: 700,
  textTransform: 'uppercase',
  letterSpacing: '0.1em',
  color: 'var(--hub-text-dim)',
};

interface RowProps {
  label: React.ReactNode;
  value: number;
  currency: Currency;
  indent?: boolean;
  total?: boolean;
  sub?: string;
  tone?: 'profit' | 'neutral';
}

/**
 * One statement line. Amounts are shown as they are: costs are listed under
 * their own heading, so they read as positive figures rather than carrying a
 * minus sign that turned an empty month into "-$0".
 */
const PLRow: React.FC<RowProps> = ({ label, value, currency, indent, total, sub, tone }) => {
  const color =
    tone === 'profit'
      ? value < 0
        ? 'var(--finance-danger)'
        : 'var(--finance-success)'
      : 'var(--hub-text)';
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: 12,
        padding: total ? '12px 24px' : '7px 24px',
        borderTop: total ? '1px solid var(--hub-border)' : 'none',
        paddingInlineStart: indent ? 40 : 24,
      }}
    >
      <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        <span
          style={{
            fontSize: total ? 14 : 13,
            fontWeight: total ? 700 : 400,
            color: total ? 'var(--hub-text)' : 'var(--hub-text-muted)',
          }}
        >
          {label}
        </span>
        {sub && <span style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{sub}</span>}
      </div>
      <span
        style={{
          fontSize: total ? 15 : 13,
          fontWeight: total ? 700 : 500,
          color,
          fontVariantNumeric: 'tabular-nums',
          whiteSpace: 'nowrap',
        }}
      >
        {formatMoney(value, currency)}
      </span>
    </div>
  );
};

const ProfitLoss: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const [period, setPeriod] = useState<Period>('month');
  const currency = useSettings((s) => s.currency);
  const {
    data: plData,
    isPending,
    isLoading,
    isError = false,
    isFetching = false,
    refetch,
  } = usePLByPeriod(period);
  const effectiveAccess = useUserStore((state) => state.user?.effectiveAccess);
  const canExport = canFinancePermission('finance:reports:export', effectiveAccess);
  const loading = Boolean(isPending || isLoading);
  const periodLabel = t(`finance.profitLoss.periods.${period}`);

  const revenue = plData?.revenue ?? 0;
  const cogs = plData?.cogs ?? 0;
  const operatingExpenses = plData?.operatingExpenses ?? 0;
  const grossProfit = revenue - cogs;
  const netProfit = grossProfit - operatingExpenses;
  const categories = plData?.expensesByCategory ?? [];
  const pendingCount = plData?.pendingExpenseCount ?? 0;

  const summary = [
    {
      id: 'revenue',
      label: t('finance.profitLoss.revenue'),
      value: formatMoney(revenue, currency),
    },
    {
      id: 'expenses',
      label: t('finance.profitLoss.expenses'),
      value: formatMoney(cogs + operatingExpenses, currency),
    },
    {
      id: 'netProfit',
      label: t('finance.profitLoss.netProfit'),
      value: formatMoney(netProfit, currency),
      color: netProfit < 0 ? 'var(--finance-danger)' : 'var(--finance-success)',
    },
    { id: 'netMargin', label: t('finance.profitLoss.netMargin'), value: pct(netProfit, revenue) },
  ];

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeNav.pl'))} />
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 12,
          flexWrap: 'wrap',
        }}
      >
        <div>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hub-text)' }}>
            {t('finance.profitLoss.title')}
          </h2>
          <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
            {t('finance.profitLoss.subtitle')}
          </p>
        </div>
        {canExport && (
          <button
            type="button"
            style={ghostBtn}
            disabled={loading || isError || isFetching || !plData}
            onClick={() =>
              exportProfitLossCsv({
                period: periodLabel,
                revenue,
                cogs,
                grossProfit,
                operatingExpenses,
                netProfit,
              })
            }
            aria-label={t('finance.profitLoss.exportAria')}
          >
            <Download size={14} />
            {t('finance.profitLoss.export')}
          </button>
        )}
      </div>

      <div
        role="tablist"
        aria-label={t('finance.profitLoss.periodLabel')}
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
        {PERIODS.map((p) => (
          <button
            type="button"
            role="tab"
            aria-selected={period === p}
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
              color: period === p ? 'var(--finance-on-accent)' : 'var(--hub-text-muted)',
              transition: 'all 0.15s',
            }}
          >
            {t(`finance.profitLoss.periods.${p}`)}
          </button>
        ))}
      </div>

      {loading ? (
        <div style={{ ...cardStyle, padding: 20 }}>
          <SkeletonRow />
        </div>
      ) : isError || !plData ? (
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <TrendingDown size={16} style={{ color: 'var(--finance-danger)' }} />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('financeErrors.generic')}
            </span>
          </div>
          <button
            type="button"
            style={{ ...ghostBtn, alignSelf: 'flex-start' }}
            disabled={isFetching}
            onClick={() => void refetch?.()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      ) : (
        <>
          {pendingCount > 0 && (
            <div
              role="status"
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 12,
                flexWrap: 'wrap',
                padding: '12px 16px',
                borderRadius: 10,
                background: 'var(--finance-warning-soft)',
                border: '1px solid var(--finance-warning-border)',
                fontSize: 13,
                color: 'var(--hub-text)',
              }}
            >
              <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Clock size={14} style={{ color: 'var(--finance-warning)', flexShrink: 0 }} />
                {t('finance.profitLoss.pendingNotice', {
                  count: pendingCount,
                  amount: formatMoney(plData.pendingExpenses, currency),
                })}
              </span>
              <Link
                to="/finance/expenses"
                style={{ fontSize: 13, fontWeight: 600, color: 'var(--finance-accent)' }}
              >
                {t('finance.profitLoss.reviewExpenses')}
              </Link>
            </div>
          )}

          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
              gap: 14,
            }}
          >
            {summary.map(({ id, label, value, color }) => (
              <div key={id} style={{ ...cardStyle, padding: 18 }}>
                <div style={{ fontSize: 12, color: 'var(--hub-text-muted)', marginBottom: 8 }}>
                  {label}
                </div>
                <div
                  style={{
                    fontSize: 22,
                    fontWeight: 700,
                    color: color ?? 'var(--hub-text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {value}
                </div>
              </div>
            ))}
          </div>

          <div style={{ ...cardStyle, overflow: 'hidden' }}>
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
              <span style={{ fontSize: 13, fontWeight: 700, color: 'var(--hub-text)' }}>
                {t('finance.profitLoss.statement')}
              </span>
              <span style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{periodLabel}</span>
            </div>

            <div style={{ padding: '8px 0' }}>
              <div style={sectionLabel}>{t('finance.profitLoss.revenue')}</div>
              <PLRow
                label={t('finance.profitLoss.paidInvoices')}
                value={revenue}
                currency={currency}
                indent
              />
            </div>

            {cogs > 0 && (
              <div style={{ borderTop: '1px solid var(--hub-border)', padding: '8px 0' }}>
                <div style={sectionLabel}>{t('finance.profitLoss.cogs')}</div>
                <PLRow
                  label={t('finance.profitLoss.directCosts')}
                  value={cogs}
                  currency={currency}
                  indent
                />
              </div>
            )}

            <PLRow
              label={t('finance.profitLoss.grossProfit')}
              value={grossProfit}
              currency={currency}
              total
              tone="profit"
              sub={t('finance.profitLoss.grossMargin', { margin: pct(grossProfit, revenue) })}
            />

            <div style={{ borderTop: '1px solid var(--hub-border)', padding: '8px 0' }}>
              <div style={sectionLabel}>{t('finance.profitLoss.operatingExpenses')}</div>
              {categories.length === 0 ? (
                <PLRow
                  label={t('finance.profitLoss.noApprovedExpenses')}
                  value={0}
                  currency={currency}
                  indent
                />
              ) : (
                categories.map((c) => (
                  <PLRow
                    key={c.categoryId ?? 'none'}
                    label={
                      <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
                        <span
                          style={{
                            width: 8,
                            height: 8,
                            borderRadius: 4,
                            background: c.color || 'var(--finance-gray)',
                          }}
                        />
                        {c.name || t('financeExpenses.uncategorized')}
                      </span>
                    }
                    value={c.amount}
                    currency={currency}
                    indent
                  />
                ))
              )}
              <PLRow
                label={t('finance.profitLoss.totalOperatingExpenses')}
                value={operatingExpenses}
                currency={currency}
                total
              />
            </div>

            <PLRow
              label={t('finance.profitLoss.netProfit')}
              value={netProfit}
              currency={currency}
              total
              tone="profit"
              sub={t('finance.profitLoss.netMarginLine', { margin: pct(netProfit, revenue) })}
            />
          </div>
          <p style={{ margin: 0, fontSize: 12, color: 'var(--hub-text-dim)' }}>
            {t('finance.profitLoss.basisNote')}
          </p>
        </>
      )}
    </motion.div>
  );
};

export default ProfitLoss;
