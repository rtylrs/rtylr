import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { Plus, Paperclip, Check } from 'lucide-react';
import { safeRatio } from '@voxire/api';
import { Expense } from '../../shared/services/financeApi';
import { useExpenses, useExpenseCategories } from './useExpenses';
import { SkeletonRow } from '../../shared/components/Skeleton';

const ACCENT = 'var(--finance-accent)';
const EXPENSE_PAGE_SIZE = 25;

const categoryColors: Record<string, string> = {
  'Food & Bev': 'var(--finance-warning)',
  Utilities: 'var(--finance-info)',
  Rent: 'var(--finance-purple)',
  Staff: 'var(--finance-accent)',
  Marketing: 'var(--finance-rose)',
  Other: 'var(--finance-neutral)',
};

function getCategoryColor(cat: string): string {
  return categoryColors[cat] ?? 'var(--finance-gray)';
}

type CategoryNameLookup = (categoryId?: number | null) => string;

import { formatMoney, type Currency } from '../../shared/lib/money';

const fmt = (n: number, currency: Currency = 'USD') => formatMoney(n, currency, 0);

const cardStyle: React.CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
  borderRadius: 12,
  padding: 20,
};

// ── CSS "pie" breakdown (horizontal stacked bar) ──────────────
const CategoryBar: React.FC<{ expenses: Expense[]; categoryName: CategoryNameLookup }> = ({
  expenses,
  categoryName,
}) => {
  const categoryTotals = expenses.reduce<Record<string, number>>((acc, exp) => {
    const name = categoryName(exp.categoryId);
    acc[name] = (acc[name] || 0) + exp.amount;
    return acc;
  }, {});
  const grandTotal = Object.values(categoryTotals).reduce((s, v) => s + v, 0);
  const entries = Object.entries(categoryTotals);
  entries.sort((a, b) => b[1] - a[1]);

  if (entries.length === 0)
    return <div style={{ height: 20, background: 'var(--hub-hover)', borderRadius: 8 }} />;

  return (
    <div>
      {/* Stacked bar */}
      <div
        style={{
          display: 'flex',
          borderRadius: 8,
          overflow: 'hidden',
          height: 20,
          marginBottom: 16,
        }}
      >
        {entries.map(([cat, val]) => (
          <div
            key={cat}
            style={{
              width: `${safeRatio(val, grandTotal) * 100}%`,
              background: getCategoryColor(cat),
              transition: 'width 0.4s ease',
            }}
            title={`${cat}: ${fmt(val)}`}
          />
        ))}
      </div>

      {/* Legend */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '10px 16px' }}>
        {entries.map(([cat, val]) => (
          <div key={cat} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span
              style={{
                width: 10,
                height: 10,
                borderRadius: 3,
                background: getCategoryColor(cat),
                flexShrink: 0,
              }}
            />
            <div style={{ minWidth: 0 }}>
              <div
                style={{
                  fontSize: 12,
                  fontWeight: 500,
                  color: 'var(--hub-text)',
                  whiteSpace: 'nowrap',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {cat}
              </div>
              <div style={{ fontSize: 11, color: 'var(--hub-text-muted)' }}>
                {fmt(val)} &middot; {(safeRatio(val, grandTotal) * 100).toFixed(1)}%
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

/** `YYYY-MM` for the month the reader is standing in, local rather than UTC. */
function currentMonthIso(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
}

const Expenses: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const {
    data,
    isLoading: loading,
    isError,
    refetch,
  } = useExpenses({
    page: 1,
    pageSize: Number.MAX_SAFE_INTEGER,
  });
  const { data: categories = [] } = useExpenseCategories();
  const allExpenses: Expense[] = data?.expenses ?? [];
  const start = (page - 1) * EXPENSE_PAGE_SIZE;
  const expenses = allExpenses.slice(start, start + EXPENSE_PAGE_SIZE);
  const hasMore = start + EXPENSE_PAGE_SIZE < allExpenses.length;

  const categoryName: CategoryNameLookup = useMemo(() => {
    const byId = new Map<number, string>(categories.map((c) => [c.id, c.name]));
    return (categoryId) =>
      categoryId != null ? (byId.get(categoryId) ?? 'Uncategorized') : 'Uncategorized';
  }, [categories]);

  // Scope the total to the month the reader is standing in. `useExpenses` walks
  // every page of `listExpenses`, and the contract carries no date filter, so
  // `allExpenses` is the entire ledger - summing it presented every prior
  // year's rent and payroll under the label "Monthly Total" in all three
  // locales, disagreeing with the P&L screen that does scope by month.
  const monthlyTotal = allExpenses.reduce(
    (sum, e) => (e.expenseDate.slice(0, 7) === currentMonthIso() ? sum + e.amount : sum),
    0
  );

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeNav.expenses'))} />
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: 'var(--hub-text)' }}>
            {t('financeExpenses.title')}
          </h2>
          {!loading && !isError && (
            <p style={{ margin: '2px 0 0', fontSize: 13, color: 'var(--hub-text-muted)' }}>
              Monthly total:{' '}
              <strong style={{ color: 'var(--hub-text)' }}>{fmt(monthlyTotal)}</strong>
            </p>
          )}
        </div>
        <button
          type="button"
          onClick={() => navigate({ to: '/finance/expenses/new' })}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            padding: '8px 16px',
            borderRadius: 8,
            background: ACCENT,
            color: 'var(--finance-white)',
            border: 'none',
            fontSize: 13,
            fontWeight: 600,
            cursor: 'pointer',
          }}
          aria-label={t('financeExpenses.addAria')}
        >
          <Plus size={15} />
          {t('financeExpenses.add')}
        </button>
      </div>

      {loading ? (
        <div style={cardStyle}>
          <SkeletonRow />
        </div>
      ) : isError ? (
        <div role="alert" style={cardStyle}>
          <p style={{ margin: '0 0 12px', fontSize: 13, color: 'var(--hub-text)' }}>
            {t('financeErrors.generic')}
          </p>
          <button
            type="button"
            onClick={() => void refetch()}
            style={{
              background: 'var(--hub-hover)',
              color: 'var(--hub-text-muted)',
              border: 'none',
              borderRadius: 8,
              padding: '7px 12px',
              fontSize: 12,
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      ) : (
        <>
          {/* Category breakdown */}
          <div style={cardStyle}>
            <h3
              style={{
                margin: '0 0 16px',
                fontSize: 14,
                fontWeight: 600,
                color: 'var(--hub-text)',
              }}
            >
              {t('financeExpenses.breakdown')}
            </h3>
            <CategoryBar expenses={allExpenses} categoryName={categoryName} />
          </div>

          {/* Expenses table */}
          <div
            style={{
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: 12,
              overflow: 'hidden',
            }}
          >
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                  {['Date', 'Description', 'Category', 'Amount', 'Receipt'].map((h) => (
                    <th
                      key={h}
                      style={{
                        padding: '10px 16px',
                        textAlign: 'start',
                        fontSize: 11,
                        fontWeight: 600,
                        textTransform: 'uppercase',
                        letterSpacing: '0.07em',
                        color: 'var(--hub-text-dim)',
                      }}
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {expenses.map((exp, i) => (
                  <tr
                    key={exp.id}
                    style={{
                      borderBottom:
                        i < expenses.length - 1 ? '1px solid var(--hub-border)' : 'none',
                    }}
                    onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--hub-hover)')}
                    onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
                  >
                    <td
                      style={{ padding: '11px 16px', fontSize: 12, color: 'var(--hub-text-muted)' }}
                    >
                      {exp.expenseDate}
                    </td>
                    <td style={{ padding: '11px 16px', fontSize: 13, color: 'var(--hub-text)' }}>
                      <div>{exp.description}</div>
                      <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{exp.id}</div>
                    </td>
                    <td style={{ padding: '11px 16px' }}>
                      {(() => {
                        const name = categoryName(exp.categoryId);
                        return (
                          <span
                            style={{
                              display: 'inline-flex',
                              alignItems: 'center',
                              gap: 5,
                              padding: '3px 10px',
                              borderRadius: 20,
                              fontSize: 11,
                              fontWeight: 600,
                              background: `${getCategoryColor(name)}18`,
                              color: getCategoryColor(name),
                            }}
                          >
                            {name}
                          </span>
                        );
                      })()}
                    </td>
                    <td
                      style={{
                        padding: '11px 16px',
                        fontSize: 13,
                        fontWeight: 600,
                        color: 'var(--finance-danger)',
                        fontVariantNumeric: 'tabular-nums',
                      }}
                    >
                      -{fmt(exp.amount)}
                    </td>
                    <td style={{ padding: '11px 16px' }}>
                      {exp.receiptUrl ? (
                        <span
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 4,
                            fontSize: 11,
                            color: 'var(--finance-accent)',
                          }}
                        >
                          <Check size={12} /> Yes
                        </span>
                      ) : (
                        <span
                          style={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 4,
                            fontSize: 11,
                            color: 'var(--hub-text-dim)',
                          }}
                        >
                          <Paperclip size={12} /> Upload
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr
                  style={{
                    borderTop: '1px solid var(--hub-border)',
                    background: 'var(--hub-hover)',
                  }}
                >
                  <td
                    colSpan={3}
                    style={{
                      padding: '10px 16px',
                      fontSize: 12,
                      fontWeight: 600,
                      color: 'var(--hub-text-muted)',
                    }}
                  >
                    {t('financeExpenses.monthlyTotal')}
                  </td>
                  <td
                    style={{
                      padding: '10px 16px',
                      fontSize: 14,
                      fontWeight: 700,
                      color: 'var(--finance-danger)',
                      fontVariantNumeric: 'tabular-nums',
                    }}
                  >
                    -{fmt(monthlyTotal)}
                  </td>
                  <td />
                </tr>
              </tfoot>
            </table>
          </div>

          {(page > 1 || hasMore) && (
            <div className="flex items-center justify-between px-1 py-3">
              <p className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                Page {page}
              </p>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page <= 1}
                  className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                  style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                >
                  {t('financeCommon.prev')}
                </button>
                <button
                  type="button"
                  onClick={() => setPage((p) => p + 1)}
                  disabled={!hasMore}
                  className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors disabled:opacity-30"
                  style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-muted)' }}
                >
                  {t('financeCommon.next')}
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </motion.div>
  );
};

export default Expenses;
