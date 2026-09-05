import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import {
  ArrowRight,
  ArrowUpRight,
  Check,
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
  ListChecks,
  Rocket,
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
// static config + localStorage (command palette, recents, checklist). Heavy
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
const CHECKLIST_KEY = 'rtylr-home-checklist';
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

interface ChecklistStep {
  id: string;
  title: string;
  description: string;
  path?: string;
  href?: string;
  external?: boolean;
}

const CHECKLIST_STEPS: ChecklistStep[] = [
  {
    id: 'invite-team',
    title: 'Invite your team',
    description: 'Add staff and control which apps they can access.',
    path: '/team/invite',
  },
  {
    id: 'open-pos',
    title: 'Open Point of Sale',
    description: 'Take your first order.',
    href: APP_URLS.pos,
    external: true,
  },
  {
    id: 'setup-menu',
    title: 'Set up your menu',
    description: 'Add items, menus, and QR codes for your storefront.',
    path: '/menus',
  },
  {
    id: 'explore-analytics',
    title: 'Explore analytics',
    description: 'Revenue, customers, and products live here — loaded on demand.',
    path: '/analytics',
  },
  {
    id: 'connect-integration',
    title: 'Connect an integration',
    description: 'API keys, webhooks, and add-ons.',
    path: '/integrations',
  },
];

function readChecklist(): string[] {
  try {
    const raw = localStorage.getItem(getTenantKey(CHECKLIST_KEY));
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter((id) => typeof id === 'string') : [];
  } catch {
    return [];
  }
}

const cardStyle: React.CSSProperties = {
  background: 'var(--dash-card)',
  border: '1px solid var(--dash-border)',
};

const AdminHome: React.FC = () => {
  const { user } = useUserStore();
  const { setOpen: setPaletteOpen } = useCommandPalette();
  const navigate = useNavigate();
  const businessName = (user?.businessName as string | undefined) || 'Welcome';

  const [recents, setRecents] = useState<RecentEntry[]>([]);
  const [done, setDone] = useState<string[]>([]);

  useEffect(() => {
    setRecents(readRecents());
    setDone(readChecklist());
  }, []);

  const persistDone = useCallback((next: string[]) => {
    setDone(next);
    try {
      localStorage.setItem(getTenantKey(CHECKLIST_KEY), JSON.stringify(next));
    } catch {
      // best-effort
    }
  }, []);

  const toggleStep = useCallback(
    (id: string) => {
      persistDone(done.includes(id) ? done.filter((d) => d !== id) : [...done, id]);
    },
    [done, persistDone]
  );

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

  const remaining = useMemo(() => CHECKLIST_STEPS.filter((s) => !done.includes(s.id)), [done]);

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      {/* Header */}
      <div className="text-center pt-4">
        <h1 className="text-2xl font-semibold tracking-tight" style={{ color: 'var(--dash-text)' }}>
          {businessName}
        </h1>
        <p className="text-xs mt-1" style={{ color: 'var(--dash-text-dim)' }}>
          {today}
        </p>
      </div>

      {/* Big search — opens the existing command palette, zero queries */}
      <button
        type="button"
        onClick={() => setPaletteOpen(true)}
        className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-left transition-colors hover:bg-[var(--dash-hover)]"
        style={{ ...cardStyle, color: 'var(--dash-text-dim)' }}
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

      {/* Three columns — Cloudflare style, all static */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Apps */}
        <section className="rounded-xl overflow-hidden" style={cardStyle}>
          <header
            className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider"
            style={{ color: 'var(--dash-text-dim)', borderBottom: '1px solid var(--dash-border)' }}
          >
            Apps
          </header>
          <ul className="divide-y" style={{ borderColor: 'var(--dash-border)' }}>
            {APP_LINKS.map((app) => {
              const Icon = app.icon;
              return (
                <li key={app.id}>
                  <button
                    type="button"
                    onClick={() => openExternal(app.name, app.href, `/apps/${app.id}`)}
                    className="w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-[var(--dash-hover)]"
                  >
                    <span
                      className="h-7 w-7 rounded-lg flex items-center justify-center shrink-0"
                      style={{ background: `${app.color}12` }}
                    >
                      <Icon size={14} style={{ color: app.color }} />
                    </span>
                    <span className="flex-1 min-w-0">
                      <span
                        className="block text-[13px] font-medium truncate"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {app.name}
                      </span>
                      <span
                        className="block text-[11px] truncate"
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
        <section className="rounded-xl overflow-hidden" style={cardStyle}>
          <header
            className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider"
            style={{ color: 'var(--dash-text-dim)', borderBottom: '1px solid var(--dash-border)' }}
          >
            Shortcuts
          </header>
          <ul className="divide-y" style={{ borderColor: 'var(--dash-border)' }}>
            {SHORTCUTS.map((s) => {
              const Icon = s.icon;
              return (
                <li key={s.path}>
                  <button
                    type="button"
                    onClick={() => goInternal(s.label, s.path)}
                    className="w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-[var(--dash-hover)]"
                  >
                    <Icon size={14} style={{ color: 'var(--dash-text-dim)' }} />
                    <span className="flex-1 min-w-0">
                      <span
                        className="block text-[13px] font-medium truncate"
                        style={{ color: 'var(--dash-text)' }}
                      >
                        {s.label}
                      </span>
                      <span
                        className="block text-[11px] truncate"
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
        <section className="rounded-xl overflow-hidden" style={cardStyle}>
          <header
            className="px-4 py-2.5 text-xs font-semibold uppercase tracking-wider flex items-center gap-1.5"
            style={{ color: 'var(--dash-text-dim)', borderBottom: '1px solid var(--dash-border)' }}
          >
            <History size={12} />
            Recents
          </header>
          {recents.length === 0 ? (
            <p
              className="px-4 py-6 text-xs leading-relaxed"
              style={{ color: 'var(--dash-text-dim)' }}
            >
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
            <ul className="divide-y" style={{ borderColor: 'var(--dash-border)' }}>
              {recents.map((r) => (
                <li key={r.path}>
                  <button
                    type="button"
                    onClick={() => goInternal(r.label, r.path)}
                    className="w-full flex items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-[var(--dash-hover)]"
                  >
                    <History size={14} style={{ color: 'var(--dash-text-dim)' }} />
                    <span
                      className="flex-1 text-[13px] font-medium truncate"
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
          <div className="px-4 py-3" style={{ borderTop: '1px solid var(--dash-border)' }}>
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

      {/* Getting started — BetterStack style, static + localStorage */}
      <section className="rounded-xl overflow-hidden" style={cardStyle}>
        <header className="px-5 py-3.5 flex items-center justify-between gap-3">
          <p
            className="text-sm font-semibold flex items-center gap-2"
            style={{ color: 'var(--dash-text)' }}
          >
            <Rocket size={14} style={{ color: 'var(--accent)' }} />
            Get the most out of rtylr
          </p>
          <p className="text-xs shrink-0" style={{ color: 'var(--dash-text-dim)' }}>
            {remaining.length === 0 ? (
              'All done 🎉'
            ) : (
              <>
                <span className="font-semibold" style={{ color: 'var(--dash-text)' }}>
                  {remaining.length}
                </span>{' '}
                of {CHECKLIST_STEPS.length} steps left
              </>
            )}
          </p>
        </header>
        <ul className="divide-y" style={{ borderColor: 'var(--dash-border)' }}>
          {CHECKLIST_STEPS.map((step) => {
            const isDone = done.includes(step.id);
            return (
              <li key={step.id} className="flex items-center gap-3 px-5 py-3">
                <button
                  type="button"
                  role="checkbox"
                  aria-checked={isDone}
                  aria-label={`Mark "${step.title}" as ${isDone ? 'not done' : 'done'}`}
                  onClick={() => toggleStep(step.id)}
                  className="h-5 w-5 rounded-full flex items-center justify-center shrink-0 transition-colors"
                  style={{
                    background: isDone ? 'var(--accent)' : 'transparent',
                    border: isDone ? '1px solid var(--accent)' : '1px dashed var(--dash-text-dim)',
                    color: '#fff',
                  }}
                >
                  {isDone && <Check size={12} />}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    if (step.href) openExternal(step.title, step.href, step.path ?? step.href);
                    else if (step.path) goInternal(step.title, step.path);
                  }}
                  className="flex-1 min-w-0 flex items-center gap-3 text-left group"
                >
                  <span className="flex-1 min-w-0">
                    <span
                      className="block text-[13px] font-medium truncate"
                      style={{
                        color: isDone ? 'var(--dash-text-dim)' : 'var(--dash-text)',
                        textDecoration: isDone ? 'line-through' : 'none',
                      }}
                    >
                      {step.title}
                    </span>
                    <span
                      className="block text-[11px] truncate"
                      style={{ color: 'var(--dash-text-dim)' }}
                    >
                      {step.description}
                    </span>
                  </span>
                  <ChevronRight
                    size={14}
                    className="shrink-0 transition-transform group-hover:translate-x-0.5"
                    style={{ color: 'var(--dash-text-dim)' }}
                  />
                </button>
              </li>
            );
          })}
        </ul>
        <div
          className="px-5 py-3 flex items-center gap-1.5 text-[11px]"
          style={{ borderTop: '1px solid var(--dash-border)', color: 'var(--dash-text-dim)' }}
        >
          <ListChecks size={12} />
          Progress is saved on this device only — no server round-trips.
        </div>
      </section>
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
