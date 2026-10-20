import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { AlertCircle, Search } from 'lucide-react';
import { useBusinessCurrency } from '../../../../../shared/hooks/queries/useBusinesses';
import { formatMoney } from '../../../../../shared/lib/money';
import { Button } from '../../../shared/components/ui';
import { useReps } from '../data/queries';
import {
  amountOf,
  type DealStatus,
  dueState,
  formatDay,
  initials,
  type LeadStatus,
} from '../data/model';
import '../sales.css';

const LEAD_TONES: Record<LeadStatus, string> = {
  new: 'blue',
  attempting: 'amber',
  connected: 'violet',
  nurture: 'teal',
  qualified: 'green',
  disqualified: 'red',
};

export const LeadStatusBadge: React.FC<{ status?: string }> = ({ status }) => {
  const { t } = useTranslation();
  const known = (status ?? 'new') as LeadStatus;
  return (
    <span className="sales-badge" data-tone={LEAD_TONES[known] ?? 'blue'}>
      {t(`crmSales.leadStatus.${known}`)}
    </span>
  );
};

const DEAL_TONES: Record<DealStatus, string> = { open: 'blue', won: 'green', lost: 'red' };

export const DealStatusBadge: React.FC<{ status?: string }> = ({ status }) => {
  const { t } = useTranslation();
  const known = (status ?? 'open') as DealStatus;
  return (
    <span className="sales-badge" data-tone={DEAL_TONES[known] ?? 'blue'}>
      {t(`crmSales.dealStatus.${known}`)}
    </span>
  );
};

/** Formats money in the business's currency for the reader's locale. */
export function useMoney(): (value?: string | number) => string {
  const currency = useBusinessCurrency();
  const { i18n } = useTranslation();
  return (value) =>
    formatMoney(typeof value === 'number' ? value : amountOf(value), currency, i18n.language);
}

/** A value with its billing period, e.g. "$500.00 / mo". */
export const Money: React.FC<{ value?: string; billing?: string; empty?: string }> = ({
  value,
  billing,
  empty = '—',
}) => {
  const { t } = useTranslation();
  const money = useMoney();
  if (value === undefined || value === '' || amountOf(value) === 0) {
    return <span className="sales-muted">{empty}</span>;
  }
  const suffix =
    billing === 'monthly' || billing === 'yearly' ? t(`crmSales.billingSuffix.${billing}`) : '';
  return (
    <span>
      {money(value)}
      {suffix && <span className="sales-muted"> {suffix}</span>}
    </span>
  );
};

/** Names of the business's reps by id. */
export function useRepNames(): Map<string, string> {
  const { data } = useReps();
  return useMemo(() => new Map((data ?? []).map((rep) => [rep.id, rep.name])), [data]);
}

export const RepName: React.FC<{ id?: string; names: Map<string, string> }> = ({ id, names }) => {
  const { t } = useTranslation();
  if (!id) return <span className="sales-muted">{t('crmSales.team.unassigned')}</span>;
  const name = names.get(id) ?? t('crmSales.team.former');
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
      <span className="sales-avatar" aria-hidden="true">
        {initials(name)}
      </span>
      <span>{name}</span>
    </span>
  );
};

/** A follow-up date, red when overdue and amber when due today. */
export const DueDate: React.FC<{ date?: string; emptyLabel?: string }> = ({ date, emptyLabel }) => {
  const { t, i18n } = useTranslation();
  const state = dueState(date);
  if (!date || !state) return <span className="sales-muted">{emptyLabel ?? '—'}</span>;
  return (
    <span className="sales-due" data-state={state}>
      {state === 'upcoming' ? formatDay(date, i18n.language) : t(`crmSales.due.${state}`)}
    </span>
  );
};

export const StatTile: React.FC<{
  label: string;
  value: React.ReactNode;
  hint?: React.ReactNode;
  warning?: boolean;
}> = ({ label, value, hint, warning }) => (
  <div className={`sales-stat${warning ? ' is-warning' : ''}`}>
    <div className="sales-stat-label">{label}</div>
    <div className="sales-stat-value">{value}</div>
    {hint !== undefined && <div className="sales-stat-hint">{hint}</div>}
  </div>
);

/** A search box that reports its value after typing pauses. */
export const SearchBox: React.FC<{
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
}> = ({ value, onChange, placeholder }) => {
  const [draft, setDraft] = useState(value);
  useEffect(() => setDraft(value), [value]);
  useEffect(() => {
    if (draft === value) return;
    const timer = window.setTimeout(() => onChange(draft), 300);
    return () => window.clearTimeout(timer);
  }, [draft, value, onChange]);
  return (
    <label className="sales-search">
      <Search size={14} aria-hidden="true" />
      <span className="sales-visually-hidden">{placeholder}</span>
      <input
        type="search"
        value={draft}
        placeholder={placeholder}
        onChange={(event) => setDraft(event.target.value)}
      />
    </label>
  );
};

export const LoadFailed: React.FC<{ title: string; onRetry: () => void }> = ({
  title,
  onRetry,
}) => {
  const { t } = useTranslation();
  return (
    <div className="sales-empty" role="alert">
      <AlertCircle size={22} style={{ color: '#ef4444' }} aria-hidden="true" />
      <h3>{title}</h3>
      <p>{t('crmSales.common.refreshAndRetry')}</p>
      <Button variant="outline" onClick={onRetry}>
        {t('crmSales.common.retry')}
      </Button>
    </div>
  );
};

export const EmptyBlock: React.FC<{
  icon: React.ComponentType<{ size?: number }>;
  title: string;
  description: string;
  children?: React.ReactNode;
}> = ({ icon: Icon, title, description, children }) => (
  <div className="sales-empty">
    <Icon size={22} />
    <h3>{title}</h3>
    <p>{description}</p>
    {children && <div className="sales-actions">{children}</div>}
  </div>
);

/** "Showing 50 of 80,000" with a load-more button while pages remain. */
export const ListFooter: React.FC<{
  shown: number;
  total?: number;
  hasMore: boolean;
  loading: boolean;
  onMore: () => void;
}> = ({ shown, total, hasMore, loading, onMore }) => {
  const { t, i18n } = useTranslation();
  const format = (value: number) => new Intl.NumberFormat(i18n.language).format(value);
  return (
    <div className="sales-footer">
      <span>
        {total !== undefined
          ? t('crmSales.common.showingOf', { shown: format(shown), total: format(total) })
          : t('crmSales.common.showing', { shown: format(shown) })}
      </span>
      {hasMore && (
        <Button variant="outline" onClick={onMore} disabled={loading}>
          {loading ? t('crmSales.common.loading') : t('crmSales.common.loadMore')}
        </Button>
      )}
    </div>
  );
};

/** Rep filter options: everyone, unassigned, then each teammate. */
export const RepFilterSelect: React.FC<{
  value: string;
  onChange: (value: string) => void;
  label: string;
}> = ({ value, onChange, label }) => {
  const { t } = useTranslation();
  const { data: reps = [] } = useReps();
  return (
    <select
      className="sales-select"
      aria-label={label}
      value={value}
      onChange={(event) => onChange(event.target.value)}
    >
      <option value="all">{t('crmSales.team.everyone')}</option>
      <option value="unassigned">{t('crmSales.team.unassigned')}</option>
      {reps.map((rep) => (
        <option key={rep.id} value={rep.id}>
          {rep.name}
        </option>
      ))}
    </select>
  );
};

/** A rep picker for forms: unassigned or a teammate. */
export const RepPicker: React.FC<{
  value?: string;
  onChange: (value: string | undefined) => void;
  label: string;
  id?: string;
}> = ({ value, onChange, label, id }) => {
  const { t } = useTranslation();
  const { data: reps = [] } = useReps();
  return (
    <select
      id={id}
      className="sales-select"
      aria-label={label}
      value={value ?? ''}
      onChange={(event) => onChange(event.target.value || undefined)}
      style={{ maxWidth: 'none', width: '100%' }}
    >
      <option value="">{t('crmSales.team.unassigned')}</option>
      {reps.map((rep) => (
        <option key={rep.id} value={rep.id}>
          {rep.name}
        </option>
      ))}
    </select>
  );
};
