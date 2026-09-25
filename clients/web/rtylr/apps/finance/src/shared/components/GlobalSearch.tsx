import React, { useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import {
  Search,
  FileText,
  Receipt,
  Users,
  LayoutDashboard,
  TrendingUp,
  Calculator,
} from 'lucide-react';
import Modal from './Modal';
import { inputStyle, ACCENT } from '../lib/tokens';
import type { Client } from './ClientSelect';
import { unwrap } from '@voxire/api';
import { listCrmCustomers } from '../lib/clients';
import { parseClientsResponse } from '../schemas/clients';

interface Result {
  id: string;
  label: string;
  sub?: string;
  to: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  group: string;
}

const PAGE_KEYS: {
  id: string;
  labelKey: string;
  to: string;
  icon: Result['icon'];
  groupKey: string;
}[] = [
  {
    id: 'p-dashboard',
    labelKey: 'nav.dashboard',
    to: '/',
    icon: LayoutDashboard,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-invoices',
    labelKey: 'nav.invoices',
    to: '/invoices',
    icon: FileText,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-new-invoice',
    labelKey: 'nav.newInvoice',
    to: '/invoices/new',
    icon: FileText,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-recurring',
    labelKey: 'nav.recurringInvoices',
    to: '/invoices/recurring',
    icon: FileText,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-quotes',
    labelKey: 'nav.quotes',
    to: '/quotes',
    icon: FileText,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-expenses',
    labelKey: 'nav.expenses',
    to: '/expenses',
    icon: Receipt,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-new-expense',
    labelKey: 'nav.newExpense',
    to: '/expenses/new',
    icon: Receipt,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-categories',
    labelKey: 'nav.expenseCategories',
    to: '/categories',
    icon: Receipt,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-bills',
    labelKey: 'nav.bills',
    to: '/bills',
    icon: Receipt,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-bank',
    labelKey: 'nav.bankFeeds',
    to: '/bank',
    icon: Receipt,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-reconcile',
    labelKey: 'nav.reconcile',
    to: '/reconcile',
    icon: Receipt,
    groupKey: 'nav.groups.pages',
  },
  { id: 'p-pl', labelKey: 'nav.pl', to: '/pl', icon: TrendingUp, groupKey: 'nav.groups.pages' },
  {
    id: 'p-balance',
    labelKey: 'nav.balanceSheet',
    to: '/reports/balance-sheet',
    icon: TrendingUp,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-cashflow',
    labelKey: 'nav.cashFlow',
    to: '/reports/cash-flow',
    icon: TrendingUp,
    groupKey: 'nav.groups.pages',
  },
  { id: 'p-tax', labelKey: 'nav.tax', to: '/tax', icon: Calculator, groupKey: 'nav.groups.pages' },
  {
    id: 'p-taxreport',
    labelKey: 'nav.taxSummary',
    to: '/reports/tax',
    icon: Calculator,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-journal',
    labelKey: 'nav.journal',
    to: '/journal',
    icon: FileText,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-accounts',
    labelKey: 'nav.chartOfAccounts',
    to: '/accounts',
    icon: FileText,
    groupKey: 'nav.groups.pages',
  },
  {
    id: 'p-settings',
    labelKey: 'nav.settings',
    to: '/settings',
    icon: FileText,
    groupKey: 'nav.groups.pages',
  },
];

interface GlobalSearchProps {
  open: boolean;
  onClose: () => void;
}

const GlobalSearch: React.FC<GlobalSearchProps> = ({ open, onClose }) => {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');
  const [index, setIndex] = useState(0);
  const [clients, setClients] = useState<Client[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    if (open) {
      setQuery('');
      setIndex(0);
      void unwrap(listCrmCustomers())
        .then((data) => setClients(parseClientsResponse(data)))
        .catch(() => {});
    }
  }, [open]);

  const results = useMemo(() => {
    const q = query.toLowerCase();
    const pages: Result[] = PAGE_KEYS.map((p) => ({
      id: p.id,
      label: t(p.labelKey),
      to: p.to,
      icon: p.icon,
      group: t(p.groupKey),
    })).filter((p) => p.label.toLowerCase().includes(q));
    const clientResults: Result[] = clients
      .filter((c) => c.name.toLowerCase().includes(q))
      .map((c) => ({
        id: `client-${c.id}`,
        label: c.name,
        sub: c.email ?? undefined,
        to: `/invoices/new?client=${c.id}`,
        icon: Users,
        group: 'Clients',
      }));
    return [...pages, ...clientResults].slice(0, 12);
  }, [query, clients, t]);

  useEffect(() => {
    setIndex(0);
  }, [results]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (results.length === 0) return;
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setIndex((i) => Math.min(i + 1, results.length - 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setIndex((i) => Math.max(i - 1, 0));
      } else if (e.key === 'Enter') {
        e.preventDefault();
        const r = results[index];
        if (r) {
          navigate({ to: r.to });
          onClose();
        }
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [open, results, index, navigate, onClose]);

  return (
    <Modal open={open} onClose={onClose} title={t('common.search')} width={560}>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '8px 10px',
          background: 'var(--hub-hover)',
          borderRadius: 10,
          marginBottom: 12,
        }}
      >
        <Search size={14} style={{ color: 'var(--hub-text-dim)' }} />
        <input
          autoFocus
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setIndex(0);
          }}
          placeholder={t('finance.search.placeholder')}
          style={{ ...inputStyle, background: 'transparent', border: 'none', padding: 0 }}
        />
      </div>
      <div style={{ maxHeight: 360, overflowY: 'auto' }}>
        {results.length === 0 && (
          <div
            style={{ padding: 30, textAlign: 'center', fontSize: 12, color: 'var(--hub-text-dim)' }}
          >
            {t('finance.search.none')}
          </div>
        )}
        {results.map((r, i) => {
          const Icon = r.icon;
          const active = i === index;
          return (
            <button
              type="button"
              key={r.id}
              onMouseEnter={() => setIndex(i)}
              onClick={() => {
                navigate({ to: r.to });
                onClose();
              }}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                padding: '10px 12px',
                borderRadius: 8,
                border: 'none',
                background: active ? 'var(--finance-accent-faint)' : 'transparent',
                color: 'var(--hub-text)',
                cursor: 'pointer',
                textAlign: 'start',
                marginBottom: 2,
              }}
            >
              <Icon size={15} style={{ color: active ? ACCENT : 'var(--hub-text-dim)' }} />
              <div style={{ display: 'flex', flexDirection: 'column', minWidth: 0 }}>
                <div style={{ fontSize: 13, fontWeight: 600 }}>{r.label}</div>
                {r.sub && <div style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>{r.sub}</div>}
              </div>
              <span
                style={{
                  marginInlineStart: 'auto',
                  fontSize: 10,
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                }}
              >
                {r.group}
              </span>
            </button>
          );
        })}
      </div>
    </Modal>
  );
};

export default GlobalSearch;
