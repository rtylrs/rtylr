import React, { useEffect, useState, useMemo } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { motion, AnimatePresence } from 'motion/react';
import {
  Search,
  Users,
  Gift,
  Megaphone,
  PieChart,
  LayoutDashboard,
  Zap,
  FileText,
  MessageSquare,
  Settings as SettingsIcon,
  Plus,
  Award,
} from 'lucide-react';
import { useListCustomers } from '../hooks/queries/useCustomers';
import { CustomerAvatar } from './crm';

interface CommandItem {
  id: string;
  label: string;
  hint?: string;
  icon?: React.ComponentType<{ size?: number }>;
  action: () => void;
  group: string;
}

export const CommandPalette: React.FC<{ open: boolean; onClose: () => void }> = ({
  open,
  onClose,
}) => {
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const [cursor, setCursor] = useState(0);
  // Only fetch the customer index while the palette is open — an eager query
  // pulled 50 customer rows on every app mount.
  const { data: fetchedCustomers } = useListCustomers({ limit: 50, offset: 0 }, open);
  const customers = useMemo(() => fetchedCustomers ?? [], [fetchedCustomers]);

  useEffect(() => {
    if (!open) setQuery('');
    if (open) setCursor(0);
  }, [open]);

  const items = useMemo<CommandItem[]>(() => {
    const base: CommandItem[] = [
      {
        id: 'nav-dash',
        label: 'Dashboard',
        icon: LayoutDashboard,
        action: () => navigate({ to: '/' }),
        group: 'Navigate',
      },
      {
        id: 'nav-cust',
        label: 'Customers',
        icon: Users,
        action: () => navigate({ to: '/customers' }),
        group: 'Navigate',
      },
      {
        id: 'nav-loyalty',
        label: 'Loyalty',
        icon: Gift,
        action: () => navigate({ to: '/loyalty' }),
        group: 'Navigate',
      },
      {
        id: 'nav-camp',
        label: 'Campaigns',
        icon: Megaphone,
        action: () => navigate({ to: '/campaigns' }),
        group: 'Navigate',
      },
      {
        id: 'nav-seg',
        label: 'Segments',
        icon: PieChart,
        action: () => navigate({ to: '/segments' }),
        group: 'Navigate',
      },
      {
        id: 'nav-auto',
        label: 'Automations',
        icon: Zap,
        action: () => navigate({ to: '/automations' }),
        group: 'Navigate',
      },
      {
        id: 'nav-tmpl',
        label: 'Templates',
        icon: FileText,
        action: () => navigate({ to: '/templates' }),
        group: 'Navigate',
      },
      {
        id: 'nav-fb',
        label: 'Feedback & NPS',
        icon: MessageSquare,
        action: () => navigate({ to: '/feedback' }),
        group: 'Navigate',
      },
      {
        id: 'nav-set',
        label: 'Settings',
        icon: SettingsIcon,
        action: () => navigate({ to: '/settings' }),
        group: 'Navigate',
      },
      {
        id: 'new-cust',
        label: 'New customer',
        icon: Plus,
        action: () => navigate({ to: '/customers/new' }),
        group: 'Create',
        hint: 'n',
      },
      {
        id: 'new-seg',
        label: 'New segment',
        icon: Plus,
        action: () => navigate({ to: '/segments/new' }),
        group: 'Create',
      },
      {
        id: 'new-camp-email',
        label: 'New email campaign',
        icon: Plus,
        action: () => navigate({ to: '/campaigns/email' }),
        group: 'Create',
      },
      {
        id: 'new-camp-sms',
        label: 'New SMS campaign',
        icon: Plus,
        action: () => navigate({ to: '/campaigns/sms' }),
        group: 'Create',
      },
      {
        id: 'new-auto',
        label: 'New automation',
        icon: Plus,
        action: () => navigate({ to: '/automations/new' }),
        group: 'Create',
      },
      {
        id: 'import',
        label: 'Import customers from CSV',
        icon: Plus,
        action: () => navigate({ to: '/customers/import' }),
        group: 'Create',
      },
      {
        id: 'loyalty-members',
        label: 'Loyalty members',
        icon: Award,
        action: () => navigate({ to: '/loyalty/members' }),
        group: 'Loyalty',
      },
      {
        id: 'loyalty-rewards',
        label: 'Loyalty rewards catalog',
        icon: Gift,
        action: () => navigate({ to: '/loyalty/rewards' }),
        group: 'Loyalty',
      },
      {
        id: 'loyalty-tiers',
        label: 'Loyalty tiers',
        icon: Award,
        action: () => navigate({ to: '/loyalty/tiers' }),
        group: 'Loyalty',
      },
    ];
    const customerItems: CommandItem[] = customers.slice(0, 20).map((c) => ({
      id: `cust-${c.id}`,
      label: c.name,
      hint: c.email,
      action: () => navigate({ to: `/customers/${c.id}` }),
      group: 'Customers',
    }));
    return [...base, ...customerItems];
  }, [navigate, customers]);

  const filtered = useMemo(() => {
    if (!query) return items;
    const q = query.toLowerCase();
    return items.filter(
      (i) => i.label.toLowerCase().includes(q) || (i.hint || '').toLowerCase().includes(q)
    );
  }, [items, query]);

  const groups = useMemo(() => {
    const out: Record<string, CommandItem[]> = {};
    filtered.forEach((i) => {
      out[i.group] = out[i.group] || [];
      out[i.group].push(i);
    });
    return out;
  }, [filtered]);

  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault();
        onClose();
      }
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setCursor((c) => Math.min(c + 1, filtered.length - 1));
      }
      if (e.key === 'ArrowUp') {
        e.preventDefault();
        setCursor((c) => Math.max(c - 1, 0));
      }
      if (e.key === 'Enter') {
        e.preventDefault();
        const item = filtered[cursor];
        if (item) {
          item.action();
          onClose();
        }
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [open, filtered, cursor, onClose]);

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0,0,0,0.5)',
            zIndex: 200,
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'center',
            paddingTop: '12vh',
            padding: '12vh 20px 20px',
          }}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.96, y: -10 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.96, y: -10 }}
            onClick={(e) => e.stopPropagation()}
            style={{
              width: '100%',
              maxWidth: '560px',
              background: 'var(--hub-card)',
              border: '1px solid var(--hub-border)',
              borderRadius: '14px',
              boxShadow: 'var(--hub-shadow)',
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
              maxHeight: '60vh',
            }}
          >
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '10px',
                padding: '14px 16px',
                borderBottom: '1px solid var(--hub-border)',
              }}
            >
              <Search size={16} style={{ color: 'var(--hub-text-muted)' }} />
              <input
                autoFocus
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setCursor(0);
                }}
                placeholder="Search customers, navigate, create..."
                style={{
                  flex: 1,
                  background: 'transparent',
                  border: 'none',
                  color: 'var(--hub-text)',
                  fontSize: '14px',
                  outline: 'none',
                }}
              />
              <kbd
                style={{
                  fontSize: '10px',
                  color: 'var(--hub-text-dim)',
                  padding: '2px 6px',
                  borderRadius: '4px',
                  border: '1px solid var(--hub-border)',
                }}
              >
                esc
              </kbd>
            </div>
            <div style={{ overflowY: 'auto', flex: 1, padding: '8px 0' }}>
              {filtered.length === 0 && (
                <div
                  style={{
                    padding: '32px',
                    textAlign: 'center',
                    fontSize: '13px',
                    color: 'var(--hub-text-muted)',
                  }}
                >
                  No results for "{query}"
                </div>
              )}
              {Object.entries(groups).map(([group, groupItems]) => (
                <div key={group}>
                  <div
                    style={{
                      padding: '6px 16px',
                      fontSize: '10px',
                      fontWeight: 700,
                      color: 'var(--hub-text-dim)',
                      textTransform: 'uppercase',
                      letterSpacing: '0.08em',
                    }}
                  >
                    {group}
                  </div>
                  {groupItems.map((item) => {
                    const idx = filtered.indexOf(item);
                    const active = idx === cursor;
                    const isCustomer = item.group === 'Customers';
                    const Icon = item.icon;
                    return (
                      <button
                        aria-label="Run command"
                        type="button"
                        key={item.id}
                        onClick={() => {
                          item.action();
                          onClose();
                        }}
                        onMouseEnter={() => setCursor(idx)}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          gap: '12px',
                          width: '100%',
                          padding: '10px 16px',
                          background: active ? 'var(--hub-hover)' : 'transparent',
                          border: 'none',
                          cursor: 'pointer',
                          color: 'var(--hub-text)',
                          fontSize: '13px',
                          textAlign: 'left',
                        }}
                      >
                        {isCustomer ? (
                          <CustomerAvatar
                            initials={item.label
                              .split(' ')
                              .map((w) => w[0])
                              .join('')
                              .slice(0, 2)
                              .toUpperCase()}
                            size={24}
                          />
                        ) : (
                          Icon && (
                            <div
                              style={{
                                width: '24px',
                                height: '24px',
                                borderRadius: '6px',
                                background: 'rgba(59,130,246,0.12)',
                                color: '#3b82f6',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                flexShrink: 0,
                              }}
                            >
                              <Icon size={12} />
                            </div>
                          )
                        )}
                        <span style={{ flex: 1 }}>{item.label}</span>
                        {item.hint && (
                          <span style={{ fontSize: '11px', color: 'var(--hub-text-dim)' }}>
                            {item.hint}
                          </span>
                        )}
                      </button>
                    );
                  })}
                </div>
              ))}
            </div>
            <div
              style={{
                padding: '8px 16px',
                borderTop: '1px solid var(--hub-border)',
                fontSize: '10px',
                color: 'var(--hub-text-dim)',
                display: 'flex',
                gap: '14px',
              }}
            >
              <span>↑↓ Navigate</span>
              <span>↵ Select</span>
              <span>esc Close</span>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};
