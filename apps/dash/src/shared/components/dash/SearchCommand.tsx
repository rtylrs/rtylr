import React, { useEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from '@tanstack/react-router';
import {
  Search,
  LayoutDashboard,
  Users,
  Shield,
  KeyRound,
  Boxes,
  Bell,
  CreditCard,
  Settings,
  Plug2,
  Webhook,
  Activity as ActivityIcon,
  UserCog,
  Package,
  MonitorSpeaker,
  UserCheck,
  Layers,
  BarChart3,
  Palette,
  QrCode,
  UtensilsCrossed,
  CommandIcon,
  ArrowRight,
} from 'lucide-react';
import { APP_URLS } from '@voxire/config';
import { useCommandPalette } from '../../stores/commandPaletteStore';

interface CommandItem {
  id: string;
  label: string;
  group: string;
  icon: React.ComponentType<{ size?: number; className?: string }>;
  action: () => void;
  keywords?: string;
}

export const SearchCommand: React.FC = () => {
  const { open, setOpen } = useCommandPalette();
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const [activeIndex, setActiveIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const items: CommandItem[] = useMemo(() => {
    const go = (path: string) => () => {
      navigate({ to: path });
      setOpen(false);
      setQuery('');
    };
    const launchApp = (href: string) => () => {
      window.open(href, '_blank', 'noopener');
      setOpen(false);
      setQuery('');
    };
    return [
      // Navigation
      {
        id: 'nav-dashboard',
        label: 'Go to Dashboard',
        group: 'Navigate',
        icon: LayoutDashboard,
        action: go('/'),
      },
      {
        id: 'nav-team',
        label: 'Team Members',
        group: 'Navigate',
        icon: Users,
        action: go('/team'),
      },
      {
        id: 'nav-invite',
        label: 'Invite Member',
        group: 'Navigate',
        icon: UserCog,
        action: go('/team/invite'),
      },
      {
        id: 'nav-permissions',
        label: 'Permissions',
        group: 'Navigate',
        icon: Shield,
        action: go('/permissions'),
      },
      { id: 'nav-roles', label: 'Roles', group: 'Navigate', icon: Shield, action: go('/roles') },
      { id: 'nav-apps', label: 'All Apps', group: 'Navigate', icon: Boxes, action: go('/apps') },
      {
        id: 'nav-activity',
        label: 'Activity Log',
        group: 'Navigate',
        icon: ActivityIcon,
        action: go('/activity'),
      },
      {
        id: 'nav-audit',
        label: 'Audit Log',
        group: 'Navigate',
        icon: ScrollText,
        action: go('/audit'),
      },
      {
        id: 'nav-notifications',
        label: 'Notifications',
        group: 'Navigate',
        icon: Bell,
        action: go('/notifications'),
      },
      {
        id: 'nav-account',
        label: 'Account',
        group: 'Navigate',
        icon: UserCog,
        action: go('/account'),
      },
      {
        id: 'nav-billing',
        label: 'Billing',
        group: 'Navigate',
        icon: CreditCard,
        action: go('/billing'),
      },
      {
        id: 'nav-integrations',
        label: 'Integrations',
        group: 'Navigate',
        icon: Plug2,
        action: go('/integrations'),
      },
      {
        id: 'nav-api-keys',
        label: 'API Keys',
        group: 'Navigate',
        icon: KeyRound,
        action: go('/api-keys'),
      },
      {
        id: 'nav-webhooks',
        label: 'Webhooks',
        group: 'Navigate',
        icon: Webhook,
        action: go('/webhooks'),
      },
      {
        id: 'nav-security',
        label: 'Security',
        group: 'Navigate',
        icon: Shield,
        action: go('/security'),
      },
      {
        id: 'nav-settings',
        label: 'Settings',
        group: 'Navigate',
        icon: Settings,
        action: go('/settings'),
      },
      {
        id: 'nav-menus',
        label: 'Menus',
        group: 'Navigate',
        icon: UtensilsCrossed,
        action: go('/menus'),
      },
      { id: 'nav-qr', label: 'QR Codes', group: 'Navigate', icon: QrCode, action: go('/qr-codes') },
      {
        id: 'nav-design',
        label: 'Design',
        group: 'Navigate',
        icon: Palette,
        action: go('/design'),
      },
      {
        id: 'nav-analytics',
        label: 'Analytics',
        group: 'Navigate',
        icon: BarChart3,
        action: go('/analytics'),
      },

      // Apps launcher
      {
        id: 'app-pos',
        label: 'Launch POS',
        group: 'Apps',
        icon: MonitorSpeaker,
        action: launchApp(APP_URLS.pos),
      },
      {
        id: 'app-erp',
        label: 'Launch ERP',
        group: 'Apps',
        icon: Package,
        action: launchApp(APP_URLS.erp),
      },
      {
        id: 'app-hr',
        label: 'Launch HR',
        group: 'Apps',
        icon: UserCheck,
        action: launchApp(APP_URLS.hr),
      },
      {
        id: 'app-crm',
        label: 'Launch CRM',
        group: 'Apps',
        icon: Users,
        action: launchApp(APP_URLS.crm),
      },
      {
        id: 'app-finance',
        label: 'Launch Finance',
        group: 'Apps',
        icon: CreditCard,
        action: launchApp(APP_URLS.finance),
      },
      {
        id: 'app-flow',
        label: 'Launch Flow',
        group: 'Apps',
        icon: Layers,
        action: launchApp(APP_URLS.flow),
      },
    ];
  }, [navigate, setOpen]);

  const filtered = useMemo(() => {
    if (!query.trim()) return items;
    const q = query.toLowerCase();
    return items.filter(
      (i) =>
        i.label.toLowerCase().includes(q) ||
        i.group.toLowerCase().includes(q) ||
        (i.keywords ?? '').toLowerCase().includes(q)
    );
  }, [items, query]);

  useEffect(() => {
    setActiveIndex(0);
  }, [query, open]);

  useEffect(() => {
    if (open) {
      setTimeout(() => inputRef.current?.focus(), 0);
    } else {
      setQuery('');
    }
  }, [open]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (!open) return;
      if (e.key === 'Escape') {
        e.preventDefault();
        setOpen(false);
      } else if (e.key === 'ArrowDown') {
        e.preventDefault();
        setActiveIndex((i) => Math.min(filtered.length - 1, i + 1));
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setActiveIndex((i) => Math.max(0, i - 1));
      } else if (e.key === 'Enter') {
        e.preventDefault();
        filtered[activeIndex]?.action();
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open, filtered, activeIndex, setOpen]);

  if (!open) return null;

  const groups = filtered.reduce<Record<string, CommandItem[]>>((acc, item) => {
    (acc[item.group] ??= []).push(item);
    return acc;
  }, {});

  let flatIndex = -1;

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] px-4">
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={() => setOpen(false)}
      />
      <div
        className="relative z-10 w-full max-w-xl rounded-xl shadow-2xl overflow-hidden"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        <div
          className="flex items-center gap-3 px-4 py-3"
          style={{ borderBottom: '1px solid var(--dash-border)' }}
        >
          <Search size={16} style={{ color: 'var(--dash-text-dim)' }} />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search apps, pages, actions..."
            className="flex-1 bg-transparent outline-none text-sm"
            style={{ color: 'var(--dash-text)' }}
          />
          <kbd
            className="hidden sm:inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-mono"
            style={{
              background: 'var(--dash-hover)',
              color: 'var(--dash-text-dim)',
              border: '1px solid var(--dash-border)',
            }}
          >
            ESC
          </kbd>
        </div>
        <div className="max-h-[60vh] overflow-y-auto p-2">
          {filtered.length === 0 && (
            <div className="text-center py-10 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
              No results for &quot;{query}&quot;
            </div>
          )}
          {Object.entries(groups).map(([group, groupItems]) => (
            <div key={group} className="mb-2 last:mb-0">
              <p
                className="px-3 py-1.5 text-[10px] font-semibold uppercase tracking-widest"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {group}
              </p>
              {groupItems.map((item) => {
                flatIndex++;
                const isActive = flatIndex === activeIndex;
                const Icon = item.icon;
                return (
                  <button
                    key={item.id}
                    className="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-sm transition-colors text-left"
                    style={{
                      background: isActive ? 'var(--dash-active)' : 'transparent',
                      color: isActive ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                    }}
                    onMouseEnter={() => setActiveIndex(flatIndex)}
                    onClick={() => item.action()}
                  >
                    <Icon size={14} className="shrink-0" />
                    <span className="flex-1 truncate">{item.label}</span>
                    {isActive && <ArrowRight size={13} style={{ color: 'var(--dash-text-dim)' }} />}
                  </button>
                );
              })}
            </div>
          ))}
        </div>
        <div
          className="flex items-center justify-between gap-3 px-4 py-2 text-[10px]"
          style={{ borderTop: '1px solid var(--dash-border)', color: 'var(--dash-text-dim)' }}
        >
          <span className="flex items-center gap-1">
            <CommandIcon size={10} />
            Command palette
          </span>
          <span className="flex items-center gap-2">
            <span className="flex items-center gap-1">
              <kbd
                className="px-1 py-0.5 rounded font-mono"
                style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
              >
                ↑↓
              </kbd>
              navigate
            </span>
            <span className="flex items-center gap-1">
              <kbd
                className="px-1 py-0.5 rounded font-mono"
                style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
              >
                ↵
              </kbd>
              select
            </span>
          </span>
        </div>
      </div>
    </div>,
    document.body
  );
};

export default SearchCommand;
