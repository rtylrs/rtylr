import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import {
  ArrowRight,
  ArrowUpRight,
  ChevronRight,
  CreditCard,
  Globe,
  Layers,
  MonitorSpeaker,
  Package,
  Search,
  UserCheck,
  Users,
  BarChart3,
  Plug2,
  History,
  LayoutGrid,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { APP_URLS, resolveAppUrl } from '@voxire/config';
import { useUserStore } from '../../../shared/stores/userStore';
import { useCommandPalette } from '../../../shared/stores/commandPaletteStore';
import { getTenantKey } from '../../../shared/utils/tenantStorage';

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

// ─── Zero-query home ──────────────────────────────────────────────────────────
// Deliberately fires NO backend queries: no KPIs, no activity feed, no team
// list, no time-off scan, no revenue/invoice aggregations. Everything below is
// static config + localStorage (command palette and recents). Heavy
// analytics live in the Insights app and per-app pages, loaded on demand.

interface AppLink {
  id: string;
  name: string;
  blurb: string;
  href: string;
  external: boolean;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  color: string;
}

const APP_LINKS: AppLink[] = [
  {
    id: 'pos',
    name: 'Point of Sale',
    blurb: 'Take orders',
    href: APP_URLS.pos,
    external: true,
    icon: MonitorSpeaker,
    color: '#4583C4',
  },
  {
    id: 'erp',
    name: 'Inventory',
    blurb: 'Stock & suppliers',
    href: APP_URLS.erp,
    external: true,
    icon: Package,
    color: '#10b981',
  },
  {
    id: 'hr',
    name: 'Human Resources',
    blurb: 'Staff & shifts',
    href: APP_URLS.hr,
    external: true,
    icon: UserCheck,
    color: '#f59e0b',
  },
  {
    id: 'crm',
    name: 'Customers',
    blurb: 'CRM & loyalty',
    href: APP_URLS.crm,
    external: true,
    icon: Users,
    color: '#8b5cf6',
  },
  {
    id: 'finance',
    name: 'Finance',
    blurb: 'Invoices & bills',
    href: APP_URLS.finance,
    external: true,
    icon: CreditCard,
    color: '#ef4444',
  },
  {
    id: 'flow',
    name: 'Flow',
    blurb: 'Projects & inbox',
    href: APP_URLS.flow,
    external: true,
    icon: Layers,
    color: '#64C5B9',
  },
];

interface Shortcut {
  label: string;
  hint: string;
  path: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
}

const SHORTCUTS: Shortcut[] = [
  { label: 'Team', hint: 'Members & roles', path: '/team', icon: Users },
  { label: 'Analytics', hint: 'Revenue & reports', path: '/analytics', icon: BarChart3 },
  { label: 'Audit log', hint: 'Activity history', path: '/audit', icon: History },
  { label: 'Integrations', hint: 'API keys & webhooks', path: '/integrations', icon: Plug2 },
  { label: 'Billing', hint: 'Plan & usage', path: '/settings/billing', icon: CreditCard },
  { label: 'All apps', hint: 'Browse everything', path: '/apps', icon: LayoutGrid },
];

interface RecentEntry {
  label: string;
  path: string;
  at: number;
}

const RECENTS_KEY = 'rtylr-home-recents';
const MAX_RECENTS = 5;

function readRecents(): RecentEntry[] {
  try {
    const raw = localStorage.getItem(getTenantKey(RECENTS_KEY));
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.slice(0, MAX_RECENTS) : [];
  } catch {
    return [];
  }
}

export function recordRecent(label: string, path: string) {
  try {
    const next = [
      { label, path, at: Date.now() },
      ...readRecents().filter((r) => r.path !== path),
    ].slice(0, MAX_RECENTS);
    localStorage.setItem(getTenantKey(RECENTS_KEY), JSON.stringify(next));
  } catch {
    // localStorage full or blocked — recents are best-effort
  }
}

const AdminHome: React.FC = () => {
  const { user } = useUserStore();
  const { setOpen: setPaletteOpen } = useCommandPalette();
  const navigate = useNavigate();
  const businessName = (user?.businessName as string | undefined) || 'Welcome';

  const [recents, setRecents] = useState<RecentEntry[]>([]);

  useEffect(() => {
    setRecents(readRecents());
  }, []);

  const goInternal = useCallback(
    (label: string, path: string) => {
      recordRecent(label, path);
      setRecents(readRecents());
      void navigate({ to: path });
    },
    [navigate]
  );

  const openExternal = useCallback((label: string, href: string, pathForRecent: string) => {
    recordRecent(label, pathForRecent);
    setRecents(readRecents());
    window.open(href, '_blank', 'noopener');
  }, []);

  return (
    <div className="mx-auto w-full max-w-5xl space-y-8 pb-12 pt-6 sm:space-y-10 sm:pt-12">
      {/* Header */}
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight" style={{ color: 'var(--dash-text)' }}>
          {businessName}
        </h1>
        <p className="text-sm" style={{ color: 'var(--dash-text-dim)' }}>
          {today}
        </p>
      </div>

      {/* Big search — opens the existing command palette, zero queries */}
      <button
        type="button"
        onClick={() => setPaletteOpen(true)}
        className="w-full flex items-center gap-3 px-4 py-3.5 rounded-xl text-left transition-colors hover:bg-[var(--dash-hover)] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--accent)]"
        style={{
          background: 'var(--dash-card)',
          border: '1px solid var(--dash-border)',
          boxShadow: '0 0 0 4px color-mix(in srgb, var(--dash-border) 35%, transparent)',
          color: 'var(--dash-text-dim)',
        }}
      >
        <Search size={16} className="shrink-0" />
        <span className="flex-1 text-sm">Search apps, pages, actions…</span>
        <kbd
          className="hidden sm:inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-mono"
          style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
        >
          ⌘K
        </kbd>
      </button>

      <div className="grid grid-cols-1 gap-8 md:grid-cols-3 md:gap-7">
        {/* Apps */}
        <section className="min-w-0">
          <header
            className="px-2 pb-3 text-sm font-medium"
            style={{ color: 'var(--dash-text-dim)' }}
          >
            Apps
          </header>
          <ul className="divide-y divide-[var(--dash-border)]">
            {APP_LINKS.map((app) => {
              const Icon = app.icon;
              return (
                <li key={app.id}>
                  <button
                    type="button"
                    onClick={() => openExternal(app.name, app.href, `/apps/${app.id}`)}
                    className="w-full flex items-center gap-3 rounded-lg px-2 py-4 text-left transition-colors hover:bg-[var(--dash-hover)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--accent)]"
                  >
                    <span
                      className="h-8 w-8 rounded-lg flex items-center justify-center shrink-0"
                      style={{ background: `${app.color}12` }}
                    >
                      <Icon size={16} style={{ color: app.color }} />
                    </span>
                    <span className="flex-1 min-w-0">
                      <span
                        className="block text-sm font-medium truncate"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {app.name}
                      </span>
                      <span
                        className="block text-xs truncate"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        {app.blurb}
                      </span>
                    </span>
                    <ArrowUpRight size={13} style={{ color: 'var(--dash-text-dim)' }} />
                  </button>
                </li>
              );
            })}
          </ul>
        </section>

        {/* Shortcuts */}
        <section className="min-w-0">
          <header
            className="px-2 pb-3 text-sm font-medium"
            style={{ color: 'var(--dash-text-dim)' }}
          >
            Shortcuts
          </header>
          <ul className="divide-y divide-[var(--dash-border)]">
            {SHORTCUTS.map((s) => {
              const Icon = s.icon;
              return (
                <li key={s.path}>
                  <button
                    type="button"
                    onClick={() => goInternal(s.label, s.path)}
                    className="w-full flex items-center gap-3 rounded-lg px-2 py-4 text-left transition-colors hover:bg-[var(--dash-hover)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--accent)]"
                  >
                    <Icon size={16} style={{ color: 'var(--dash-text-dim)' }} />
                    <span className="flex-1 min-w-0">
                      <span
                        className="block text-sm font-medium truncate"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {s.label}
                      </span>
                      <span
                        className="block text-xs truncate"
                        style={{ color: 'var(--dash-text-dim)' }}
                      >
                        {s.hint}
                      </span>
                    </span>
                    <ChevronRight size={13} style={{ color: 'var(--dash-text-dim)' }} />
                  </button>
                </li>
              );
            })}
          </ul>
        </section>

        {/* Recents — localStorage only */}
        <section className="min-w-0">
          <header
            className="px-2 pb-3 text-sm font-medium flex items-center gap-1.5"
            style={{ color: 'var(--dash-text-dim)' }}
          >
            <History size={12} />
            Recents
          </header>
          {recents.length === 0 ? (
            <p className="px-2 py-4 text-sm leading-6" style={{ color: 'var(--dash-text-dim)' }}>
              Pages you open will show up here. Press{' '}
              <kbd
                className="px-1 py-0.5 rounded font-mono"
                style={{ background: 'var(--dash-hover)', border: '1px solid var(--dash-border)' }}
              >
                ⌘K
              </kbd>{' '}
              to jump anywhere.
            </p>
          ) : (
            <ul className="divide-y divide-[var(--dash-border)]">
              {recents.map((r) => (
                <li key={r.path}>
                  <button
                    type="button"
                    onClick={() => goInternal(r.label, r.path)}
                    className="w-full flex items-center gap-3 rounded-lg px-2 py-4 text-left transition-colors hover:bg-[var(--dash-hover)] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--accent)]"
                  >
                    <History size={14} style={{ color: 'var(--dash-text-dim)' }} />
                    <span
                      className="flex-1 text-sm font-medium truncate"
                      style={{ color: 'var(--dash-text)' }}
                    >
                      {r.label}
                    </span>
                    <ChevronRight size={13} style={{ color: 'var(--dash-text-dim)' }} />
                  </button>
                </li>
              ))}
            </ul>
          )}
          <div className="mx-2 mt-3 py-4" style={{ borderTop: '1px solid var(--dash-border)' }}>
            <button
              type="button"
              onClick={() => openExternal('Insights', APP_URLS.insights, '/analytics')}
              className="flex items-center gap-1.5 text-xs font-medium"
              style={{ color: 'var(--accent)' }}
            >
              <Globe size={12} />
              Analytics live in Insights
              <ArrowRight size={12} />
            </button>
          </div>
        </section>
      </div>
    </div>
  );
};

// ─── Staff (fallback) ───────────────────────────────────────────────────────

/**
 * The apps a staff member can be sent to from here.
 *
 * Also the guard for `resolveAppUrl`: `APP_URLS` has no entry for a string
 * that is not an app id, so resolving one would produce `undefined` and an
 * `href="undefined"`. Anything not on this list is not offered.
 */
const LAUNCHABLE_APPS = ['pos', 'erp', 'hr', 'crm', 'finance', 'flow', 'insights'] as const;

const isLaunchable = (app: string): app is (typeof LAUNCHABLE_APPS)[number] =>
  (LAUNCHABLE_APPS as readonly string[]).includes(app);

const StaffDashboard: React.FC<{ userApps: string[] }> = ({ userApps }) => {
  // Memoised because it is a useEffect dependency: a fresh array every render
  // would re-run the redirect on every render.
  const launchable = React.useMemo(() => userApps.filter(isLaunchable), [userApps]);

  React.useEffect(() => {
    // resolveAppUrl, not a hostname built by hand. APP_URLS already answers
    // this per environment, and resolveAppUrl layers self-host's runtime
    // config on top; interpolating rtylr.com sent a local session, and every
    // self-hosted one, to the hosted product. This runs in an effect with no
    // click involved, so a staff member with one app was redirected off the
    // moment they landed on Dash.
    const [firstApp] = launchable;
    if (firstApp) window.location.href = resolveAppUrl(firstApp);
  }, [launchable]);

  return (
    <div className="max-w-xl mx-auto py-12 text-center space-y-4">
      <h2 className="text-xl font-bold" style={{ color: 'var(--dash-text)' }}>
        Welcome to rtylr
      </h2>
      <p className="text-sm" style={{ color: 'var(--dash-text-muted)' }}>
        Select an app below to access your workspace.
      </p>
      <div className="flex flex-wrap justify-center gap-3 pt-2">
        {launchable.length > 0 ? (
          launchable.map((app) => (
            <a
              key={app}
              href={resolveAppUrl(app)}
              className="px-4 py-2 rounded-lg text-xs font-semibold uppercase tracking-wider transition-colors hover:bg-[var(--dash-hover)]"
              style={{
                background: 'var(--dash-card)',
                border: '1px solid var(--dash-border)',
                color: 'var(--dash-text)',
              }}
            >
              Open {app}
            </a>
          ))
        ) : (
          <p className="text-xs" style={{ color: 'var(--dash-text-dim)' }}>
            No apps assigned to your staff account. Please contact your workspace administrator.
          </p>
        )}
      </div>
    </div>
  );
};

// ─── Root ───────────────────────────────────────────────────────────────────

const VendorDashboard: React.FC = () => {
  const { user } = useUserStore();
  const role = (user?.role ?? '').toLowerCase();
  const isAdmin = user?.effectiveAccess
    ? user.effectiveAccess.isOwner || user.effectiveAccess.isAdmin
    : role === 'admin' || role === 'owner';
  const userApps: string[] = user?.effectiveAccess
    ? user.effectiveAccess.apps
        .filter((app) => app.licensed && app.hasAccess)
        .map((app) => app.appId)
    : (user?.apps ?? []);
  if (!isAdmin) return <StaffDashboard userApps={userApps} />;
  return (
    <>
      <HelmetContainer title="Dashboard · rtylr" />
      <AdminHome />
    </>
  );
};

export default VendorDashboard;
