import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { Link, useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import {
  Check,
  ChevronLeft,
  ChevronRight,
  Download,
  Paperclip,
  Plus,
  Receipt,
  Search,
  X,
} from 'lucide-react';
import { safeRatio } from '@voxire/api';
import { Expense, ExpenseCategory } from '../../shared/services/financeApi';
import { useExpenses, useExpenseCategories } from './useExpenses';
import { useApproveExpense, useRejectExpense } from '../../shared/hooks/queries/useExpenses';
import PageHeader from '../../shared/components/PageHeader';
import { SkeletonRow } from '../../shared/components/Skeleton';
import { cardStyle, ghostBtn, inputStyle, primaryBtn } from '../../shared/lib/tokens';
import { formatMoney, type Currency } from '../../shared/lib/money';
import { formatDate, formatDateShort, monthKey } from '../../shared/lib/date';
import { useSettings } from '../../shared/lib/settings';
import { canFinancePermission } from '../../shared/lib/financePermissions';
import { exportExpensesCsv } from '../../shared/lib/exports';
import { expenseStatusStyle } from './expenseDisplay';

const PAGE_SIZE = 25;
const UNCATEGORIZED_COLOR = 'var(--finance-gray)';
const STATUSES = ['pending', 'approved', 'rejected'] as const;

/** Shifts a `YYYY-MM` key by whole months. */
function shiftMonth(key: string, delta: number): string {
  const [year, month] = key.split('-').map(Number);
  return monthKey(new Date(year, month - 1 + delta, 1));
}

function monthTitle(key: string): string {
  const [year, month] = key.split('-').map(Number);
  return formatDate(new Date(year, month - 1, 1), { month: 'long', year: 'numeric' });
}

/** Newest first, so the month reads top-down from its latest spend. */
function byNewest(a: Expense, b: Expense): number {
  if (a.expenseDate !== b.expenseDate) return a.expenseDate < b.expenseDate ? 1 : -1;
  return Number(b.id) - Number(a.id);
}

/**
 * Sums amounts per currency and renders them home currency first. A single
 * figure across currencies would add euros to dollars.
 */
function formatTotal(rows: Expense[], home: Currency): string {
  const totals = new Map<string, number>();
  for (const row of rows) {
    const currency = row.currency || home;
    totals.set(currency, (totals.get(currency) ?? 0) + row.amount);
  }
  if (totals.size === 0) return formatMoney(0, home);
  return [...totals]
    .sort(([a], [b]) => (a === home ? -1 : b === home ? 1 : a.localeCompare(b)))
    .map(([currency, amount]) => formatMoney(amount, currency as Currency))
    .join(' + ');
}

const thStyle: React.CSSProperties = {
  padding: '10px 16px',
  textAlign: 'start',
  fontSize: 11,
  fontWeight: 600,
  textTransform: 'uppercase',
  letterSpacing: '0.07em',
  color: 'var(--hub-text-dim)',
  whiteSpace: 'nowrap',
};

const tdStyle: React.CSSProperties = { padding: '12px 16px', fontSize: 13 };

const smallBtn: React.CSSProperties = { ...ghostBtn, padding: '4px 8px', fontSize: 12 };

const SummaryCard: React.FC<{
  label: string;
  value: string;
  hint: string;
  onClick?: () => void;
  active?: boolean;
}> = ({ label, value, hint, onClick, active }) => {
  const body = (
    <>
      <div style={{ fontSize: 12, color: 'var(--hub-text-muted)' }}>{label}</div>
      <div
        style={{
          marginTop: 6,
          fontSize: 22,
          fontWeight: 700,
          color: 'var(--hub-text)',
          fontVariantNumeric: 'tabular-nums',
        }}
      >
        {value}
      </div>
      <div style={{ marginTop: 2, fontSize: 12, color: 'var(--hub-text-dim)' }}>{hint}</div>
    </>
  );
  const style: React.CSSProperties = {
    ...cardStyle,
    padding: '16px 20px',
    textAlign: 'start',
    borderColor: active ? 'var(--finance-accent)' : 'var(--hub-border)',
  };
  return onClick ? (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      style={{ ...style, cursor: 'pointer' }}
    >
      {body}
    </button>
  ) : (
    <div style={style}>{body}</div>
  );
};

const CategoryChip: React.FC<{ name: string; color: string }> = ({ name, color }) => (
  <span
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6,
      fontSize: 12,
      color: 'var(--hub-text)',
      whiteSpace: 'nowrap',
    }}
  >
    <span style={{ width: 8, height: 8, borderRadius: 4, background: color, flexShrink: 0 }} />
    {name}
  </span>
);

interface CategorySlice {
  id: string;
  name: string;
  color: string;
  amount: number;
}

const CategoryBreakdown: React.FC<{ slices: CategorySlice[]; currency: Currency }> = ({
  slices,
  currency,
}) => {
  const { t } = useTranslation('financeCommon');
  const total = slices.reduce((sum, s) => sum + s.amount, 0);
  return (
    <div style={{ ...cardStyle, padding: 20, flex: '1 1 260px', alignSelf: 'flex-start' }}>
      <h3 style={{ margin: '0 0 14px', fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
        {t('financeExpenses.byCategory')}
      </h3>
      <ul style={{ listStyle: 'none', margin: 0, padding: 0, display: 'grid', gap: 14 }}>
        {slices.map((slice) => {
          const share = safeRatio(slice.amount, total) * 100;
          return (
            <li key={slice.id}>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'baseline',
                  justifyContent: 'space-between',
                  gap: 12,
                }}
              >
                <CategoryChip name={slice.name} color={slice.color} />
                <span
                  style={{
                    fontSize: 12,
                    fontWeight: 600,
                    color: 'var(--hub-text)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {formatMoney(slice.amount, currency)}
                  <span
                    style={{ marginInlineStart: 6, fontWeight: 400, color: 'var(--hub-text-dim)' }}
                  >
                    {Math.round(share)}%
                  </span>
                </span>
              </div>
              <div
                style={{
                  marginTop: 6,
                  height: 6,
                  borderRadius: 3,
                  background: 'var(--hub-hover)',
                  overflow: 'hidden',
                }}
              >
                <div
                  style={{
                    width: `${share}%`,
                    height: '100%',
                    borderRadius: 3,
                    background: slice.color,
                  }}
                />
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
};

const Expenses: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const navigate = useNavigate();
  const homeCurrency = useSettings((s) => s.currency);
  const canApprove = canFinancePermission('finance:expenses:approve');
  const approve = useApproveExpense();
  const reject = useRejectExpense();
  const [month, setMonth] = useState(() => monthKey(new Date()));
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const {
    data,
    isLoading: loading,
    isError,
    refetch,
  } = useExpenses({ page: 1, pageSize: Number.MAX_SAFE_INTEGER });
  const { data: categories = [] } = useExpenseCategories();

  const categoryById = useMemo(
    () => new Map<number, ExpenseCategory>(categories.map((c) => [c.id, c])),
    [categories]
  );
  const categoryOf = (categoryId?: number | null): { name: string; color: string } => {
    const category = categoryId != null ? categoryById.get(categoryId) : undefined;
    return category
      ? { name: category.name, color: category.color || UNCATEGORIZED_COLOR }
      : { name: t('financeExpenses.uncategorized'), color: UNCATEGORIZED_COLOR };
  };
  const paymentLabel = (method?: string | null): string =>
    method ? t(`financeExpenses.paymentMethods.${method}`, { defaultValue: method }) : '';
  const statusLabel = (status: string): string =>
    t(`financeExpenses.status.${status}`, { defaultValue: status });

  const allExpenses = useMemo<Expense[]>(() => data?.expenses ?? [], [data]);
  // The contract has no date filter, so the whole ledger arrives and every
  // figure on this page is scoped to the month in view here. The total, the
  // breakdown and the rows used to disagree: the total was this month, the
  // breakdown was all time and the table was every row ever.
  const monthRows = useMemo(
    () => allExpenses.filter((e) => e.expenseDate.slice(0, 7) === month).sort(byNewest),
    [allExpenses, month]
  );
  const counted = monthRows.filter((e) => e.status !== 'rejected');
  const pending = monthRows.filter((e) => e.status === 'pending');
  const missingReceipts = counted.filter((e) => !e.receiptUrl).length;

  const slices = useMemo(() => {
    const byCategory = new Map<string, CategorySlice>();
    for (const e of counted) {
      if ((e.currency || homeCurrency) !== homeCurrency) continue;
      const key = e.categoryId != null ? String(e.categoryId) : 'none';
      const slice = byCategory.get(key) ?? { id: key, ...categoryOf(e.categoryId), amount: 0 };
      slice.amount += e.amount;
      byCategory.set(key, slice);
    }
    return [...byCategory.values()].sort((a, b) => b.amount - a.amount);
    // categoryOf only reads categoryById and t, both listed.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [counted, homeCurrency, categoryById, t]);

  // Search and the filters narrow the table only; the cards above keep
  // describing the whole month so they do not jump while typing.
  const query = search.trim().toLowerCase();
  const visibleRows = monthRows.filter((e) => {
    if (statusFilter !== 'all' && e.status !== statusFilter) return false;
    if (categoryFilter === 'none' && e.categoryId != null) return false;
    if (
      categoryFilter !== 'all' &&
      categoryFilter !== 'none' &&
      String(e.categoryId) !== categoryFilter
    )
      return false;
    if (!query) return true;
    return [e.description, e.vendor ?? '', e.notes ?? '', categoryOf(e.categoryId).name]
      .join(' ')
      .toLowerCase()
      .includes(query);
  });
  const filtered = query !== '' || statusFilter !== 'all' || categoryFilter !== 'all';

  const start = (page - 1) * PAGE_SIZE;
  const rows = visibleRows.slice(start, start + PAGE_SIZE);
  const hasMore = start + PAGE_SIZE < visibleRows.length;
  const isCurrentMonth = month === monthKey(new Date());
  const monthName = monthTitle(month);

  const goToMonth = (next: string): void => {
    setMonth(next);
    setPage(1);
  };
  const resetFilters = (): void => {
    setSearch('');
    setStatusFilter('all');
    setCategoryFilter('all');
    setPage(1);
  };
  const addExpense = (): void => void navigate({ to: '/finance/expenses/new' });
  const exportMonth = (): void =>
    exportExpensesCsv(
      `expenses-${month}.csv`,
      visibleRows.map((e) => ({
        date: e.expenseDate,
        description: e.description,
        vendor: e.vendor ?? '',
        category: categoryOf(e.categoryId).name,
        paymentMethod: paymentLabel(e.paymentMethod),
        status: statusLabel(e.status),
        amount: e.amount,
        currency: e.currency || homeCurrency,
        receiptUrl: e.receiptUrl ?? '',
      }))
    );

  const toolbar = (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexWrap: 'wrap',
        gap: 12,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <button
          type="button"
          style={{ ...ghostBtn, padding: 7 }}
          onClick={() => goToMonth(shiftMonth(month, -1))}
          aria-label={t('financeExpenses.previousMonth')}
        >
          <ChevronLeft size={15} className="rtl:rotate-180" />
        </button>
        <h3
          aria-live="polite"
          style={{
            margin: 0,
            minWidth: 150,
            textAlign: 'center',
            fontSize: 15,
            fontWeight: 600,
            color: 'var(--hub-text)',
          }}
        >
          {monthName}
        </h3>
        <button
          type="button"
          style={{ ...ghostBtn, padding: 7 }}
          onClick={() => goToMonth(shiftMonth(month, 1))}
          aria-label={t('financeExpenses.nextMonth')}
        >
          <ChevronRight size={15} className="rtl:rotate-180" />
        </button>
        {!isCurrentMonth && (
          <button
            type="button"
            style={{ ...ghostBtn, padding: '6px 12px', fontSize: 12 }}
            onClick={() => goToMonth(monthKey(new Date()))}
          >
            {t('financeExpenses.thisMonth')}
          </button>
        )}
      </div>
      {monthRows.length > 0 && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
          <div style={{ position: 'relative' }}>
            <Search
              size={13}
              style={{
                position: 'absolute',
                insetInlineStart: 10,
                top: '50%',
                transform: 'translateY(-50%)',
                color: 'var(--hub-text-dim)',
              }}
            />
            <input
              type="search"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              placeholder={t('financeExpenses.searchPlaceholder')}
              aria-label={t('financeExpenses.searchPlaceholder')}
              style={{ ...inputStyle, width: 200, paddingInlineStart: 30 }}
            />
          </div>
          <select
            value={statusFilter}
            onChange={(e) => {
              setStatusFilter(e.target.value);
              setPage(1);
            }}
            aria-label={t('financeCommon.status')}
            style={{ ...inputStyle, width: 'auto' }}
          >
            <option value="all">{t('financeExpenses.allStatuses')}</option>
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {statusLabel(s)}
              </option>
            ))}
          </select>
          <select
            value={categoryFilter}
            onChange={(e) => {
              setCategoryFilter(e.target.value);
              setPage(1);
            }}
            aria-label={t('financeCommon.category')}
            style={{ ...inputStyle, width: 'auto', maxWidth: 180 }}
          >
            <option value="all">{t('financeExpenses.allCategories')}</option>
            {categories.map((c) => (
              <option key={c.id} value={String(c.id)}>
                {c.name}
              </option>
            ))}
            <option value="none">{t('financeExpenses.uncategorized')}</option>
          </select>
          <button
            type="button"
            style={ghostBtn}
            onClick={exportMonth}
            disabled={visibleRows.length === 0}
          >
            <Download size={13} /> {t('financeExpenses.exportCsv')}
          </button>
        </div>
      )}
    </div>
  );

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeNav.expenses'))} />
      <PageHeader
        title={t('financeExpenses.title')}
        subtitle={t('financeExpenses.subtitle')}
        actions={
          <button
            type="button"
            onClick={addExpense}
            style={primaryBtn}
            aria-label={t('financeExpenses.addAria')}
          >
            <Plus size={15} />
            {t('financeExpenses.add')}
          </button>
        }
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
          <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text)' }}>
            {t('financeErrors.generic')}
          </p>
          <button
            type="button"
            style={{ ...ghostBtn, alignSelf: 'flex-start' }}
            onClick={() => void refetch()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      ) : allExpenses.length === 0 ? (
        <EmptyState
          icon={Receipt}
          title={t('financeExpenses.emptyTitle')}
          description={t('financeExpenses.emptyDescription')}
          action={{ label: t('financeExpenses.add'), onClick: addExpense }}
        />
      ) : (
        <>
          {toolbar}

          {monthRows.length === 0 ? (
            <EmptyState
              icon={Receipt}
              title={t('financeExpenses.emptyMonthTitle', { month: monthName })}
              description={t('financeExpenses.emptyMonthDescription')}
              action={{ label: t('financeExpenses.add'), onClick: addExpense }}
            />
          ) : (
            <>
              <div
                style={{
                  display: 'grid',
                  gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
                  gap: 12,
                }}
              >
                <SummaryCard
                  label={t('financeExpenses.spent')}
                  value={formatTotal(counted, homeCurrency)}
                  hint={t('financeExpenses.expenseCount', { count: counted.length })}
                />
                <SummaryCard
                  label={t('financeExpenses.awaitingApproval')}
                  value={formatTotal(pending, homeCurrency)}
                  hint={
                    pending.length > 0
                      ? t('financeExpenses.awaitingApprovalHint', { count: pending.length })
                      : t('financeExpenses.expenseCount', { count: 0 })
                  }
                  active={statusFilter === 'pending'}
                  onClick={
                    pending.length > 0
                      ? () => {
                          setStatusFilter(statusFilter === 'pending' ? 'all' : 'pending');
                          setPage(1);
                        }
                      : undefined
                  }
                />
                <SummaryCard
                  label={t('financeExpenses.missingReceipts')}
                  value={String(missingReceipts)}
                  hint={t('financeExpenses.missingReceiptsHint', { count: counted.length })}
                />
              </div>

              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 20 }}>
                <div style={{ ...cardStyle, flex: '999 1 560px', minWidth: 0, overflowX: 'auto' }}>
                  {visibleRows.length === 0 ? (
                    <div
                      style={{
                        padding: 24,
                        display: 'flex',
                        flexDirection: 'column',
                        alignItems: 'center',
                        gap: 10,
                        fontSize: 13,
                        color: 'var(--hub-text-muted)',
                      }}
                    >
                      {t('financeExpenses.noMatches')}
                      <button type="button" style={smallBtn} onClick={resetFilters}>
                        {t('financeExpenses.clearFilters')}
                      </button>
                    </div>
                  ) : (
                    <table style={{ width: '100%', minWidth: 640, borderCollapse: 'collapse' }}>
                      <thead>
                        <tr style={{ borderBottom: '1px solid var(--hub-border)' }}>
                          <th style={thStyle}>{t('financeCommon.date')}</th>
                          <th style={thStyle}>{t('financeCommon.description')}</th>
                          <th style={thStyle}>{t('financeCommon.category')}</th>
                          <th style={thStyle}>{t('financeCommon.status')}</th>
                          <th style={{ ...thStyle, textAlign: 'end' }}>
                            {t('financeCommon.amount')}
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {rows.map((exp, i) => {
                          const category = categoryOf(exp.categoryId);
                          const rejected = exp.status === 'rejected';
                          const details = [exp.vendor, paymentLabel(exp.paymentMethod)]
                            .filter(Boolean)
                            .join(' · ');
                          const reviewing =
                            (approve.isPending && approve.variables === exp.id) ||
                            (reject.isPending && reject.variables === exp.id);
                          return (
                            <tr
                              key={exp.id}
                              onClick={(e) => {
                                if ((e.target as HTMLElement).closest('a, button')) return;
                                void navigate({
                                  to: '/finance/expenses/$id',
                                  params: { id: String(exp.id) },
                                });
                              }}
                              style={{
                                cursor: 'pointer',
                                borderBottom:
                                  i < rows.length - 1 ? '1px solid var(--hub-border)' : 'none',
                              }}
                              onMouseEnter={(e) =>
                                (e.currentTarget.style.background = 'var(--hub-hover)')
                              }
                              onMouseLeave={(e) =>
                                (e.currentTarget.style.background = 'transparent')
                              }
                            >
                              <td
                                style={{
                                  ...tdStyle,
                                  fontSize: 12,
                                  color: 'var(--hub-text-muted)',
                                  whiteSpace: 'nowrap',
                                }}
                              >
                                {formatDateShort(exp.expenseDate)}
                              </td>
                              <td style={tdStyle}>
                                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                                  <Link
                                    to="/finance/expenses/$id"
                                    params={{ id: String(exp.id) }}
                                    style={{
                                      color: 'var(--hub-text)',
                                      fontWeight: 500,
                                      textDecoration: 'none',
                                    }}
                                  >
                                    {exp.description}
                                  </Link>
                                  {exp.receiptUrl && (
                                    <Paperclip
                                      size={12}
                                      style={{ color: 'var(--hub-text-dim)', flexShrink: 0 }}
                                      aria-label={t('financeExpenses.receiptAttached')}
                                    />
                                  )}
                                </div>
                                {details && (
                                  <div
                                    style={{
                                      marginTop: 2,
                                      fontSize: 12,
                                      color: 'var(--hub-text-dim)',
                                    }}
                                  >
                                    {details}
                                  </div>
                                )}
                              </td>
                              <td style={tdStyle}>
                                <CategoryChip name={category.name} color={category.color} />
                              </td>
                              <td style={tdStyle}>
                                {exp.status === 'pending' && canApprove ? (
                                  <div style={{ display: 'flex', gap: 6 }}>
                                    <button
                                      type="button"
                                      style={{ ...smallBtn, color: 'var(--finance-success)' }}
                                      disabled={reviewing}
                                      onClick={() => approve.mutate(exp.id)}
                                      aria-label={t('financeExpenses.approveAria', {
                                        description: exp.description,
                                      })}
                                    >
                                      <Check size={12} /> {t('financeExpenses.approve')}
                                    </button>
                                    <button
                                      type="button"
                                      style={{ ...smallBtn, color: 'var(--finance-danger)' }}
                                      disabled={reviewing}
                                      onClick={() => reject.mutate(exp.id)}
                                      aria-label={t('financeExpenses.rejectAria', {
                                        description: exp.description,
                                      })}
                                    >
                                      <X size={12} /> {t('financeExpenses.reject')}
                                    </button>
                                  </div>
                                ) : (
                                  <span
                                    style={{
                                      padding: '3px 10px',
                                      borderRadius: 20,
                                      fontSize: 11,
                                      fontWeight: 600,
                                      whiteSpace: 'nowrap',
                                      ...(expenseStatusStyle[exp.status] ??
                                        expenseStatusStyle.pending),
                                    }}
                                  >
                                    {statusLabel(exp.status)}
                                  </span>
                                )}
                              </td>
                              <td
                                style={{
                                  ...tdStyle,
                                  textAlign: 'end',
                                  fontWeight: 600,
                                  whiteSpace: 'nowrap',
                                  fontVariantNumeric: 'tabular-nums',
                                  color: rejected ? 'var(--hub-text-dim)' : 'var(--hub-text)',
                                  textDecoration: rejected ? 'line-through' : 'none',
                                }}
                              >
                                {formatMoney(
                                  exp.amount,
                                  (exp.currency || homeCurrency) as Currency
                                )}
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  )}

                  {(page > 1 || hasMore) && (
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'flex-end',
                        gap: 8,
                        padding: '10px 16px',
                        borderTop: '1px solid var(--hub-border)',
                      }}
                    >
                      <button
                        type="button"
                        onClick={() => setPage((p) => Math.max(1, p - 1))}
                        disabled={page <= 1}
                        style={smallBtn}
                      >
                        {t('financeCommon.prev')}
                      </button>
                      <button
                        type="button"
                        onClick={() => setPage((p) => p + 1)}
                        disabled={!hasMore}
                        style={smallBtn}
                      >
                        {t('financeCommon.next')}
                      </button>
                    </div>
                  )}
                </div>

                {slices.length > 0 && <CategoryBreakdown slices={slices} currency={homeCurrency} />}
              </div>
              {filtered && (
                <p style={{ margin: 0, fontSize: 12, color: 'var(--hub-text-dim)' }}>
                  {t('financeExpenses.showingFiltered', {
                    shown: visibleRows.length,
                    total: monthRows.length,
                  })}
                </p>
              )}
            </>
          )}
        </>
      )}
    </motion.div>
  );
};

export default Expenses;
