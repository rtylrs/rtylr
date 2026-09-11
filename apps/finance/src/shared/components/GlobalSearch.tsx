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

const PAGES: Result[] = [
  { id: 'p-dashboard', label: 'Dashboard', to: '/', icon: LayoutDashboard, group: 'Pages' },
  { id: 'p-invoices', label: 'Invoices', to: '/invoices', icon: FileText, group: 'Pages' },
  {
    id: 'p-new-invoice',
    label: 'New invoice',
    to: '/invoices/new',
    icon: FileText,
    group: 'Pages',
  },
  {
    id: 'p-recurring',
    label: 'Recurring invoices',
    to: '/invoices/recurring',
    icon: FileText,
    group: 'Pages',
  },
  { id: 'p-quotes', label: 'Quotes', to: '/quotes', icon: FileText, group: 'Pages' },
  { id: 'p-expenses', label: 'Expenses', to: '/expenses', icon: Receipt, group: 'Pages' },
  { id: 'p-new-expense', label: 'New expense', to: '/expenses/new', icon: Receipt, group: 'Pages' },
  {
    id: 'p-categories',
    label: 'Expense categories',
    to: '/categories',
    icon: Receipt,
    group: 'Pages',
  },
  { id: 'p-bills', label: 'Bills', to: '/bills', icon: Receipt, group: 'Pages' },
  { id: 'p-bank', label: 'Bank feeds', to: '/bank', icon: Receipt, group: 'Pages' },
  { id: 'p-reconcile', label: 'Reconcile', to: '/reconcile', icon: Receipt, group: 'Pages' },
  { id: 'p-pl', label: 'Profit & Loss', to: '/pl', icon: TrendingUp, group: 'Pages' },
  {
    id: 'p-balance',
    label: 'Balance sheet',
    to: '/reports/balance-sheet',
    icon: TrendingUp,
    group: 'Pages',
  },
  {
    id: 'p-cashflow',
    label: 'Cash flow',
    to: '/reports/cash-flow',
    icon: TrendingUp,
    group: 'Pages',
  },
  { id: 'p-tax', label: 'Tax', to: '/tax', icon: Calculator, group: 'Pages' },
  { id: 'p-taxreport', label: 'Tax summary', to: '/reports/tax', icon: Calculator, group: 'Pages' },
  { id: 'p-journal', label: 'Journal', to: '/journal', icon: FileText, group: 'Pages' },
  { id: 'p-accounts', label: 'Chart of accounts', to: '/accounts', icon: FileText, group: 'Pages' },
  { id: 'p-settings', label: 'Settings', to: '/settings', icon: FileText, group: 'Pages' },
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
    const pages = PAGES.filter((p) => p.label.toLowerCase().includes(q));
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
  }, [query, clients]);

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
