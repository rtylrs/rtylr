import React from 'react';
import { Link, Outlet, useNavigate, useRouterState } from '@tanstack/react-router';
import {
  User,
  Building2,
  Globe,
  Mail,
  CreditCard,
  Puzzle,
  ArrowLeft,
  Briefcase,
  Layers,
  Sparkles,
} from 'lucide-react';
import { cn } from '../../../../../shared/lib/utils';
import { useUserStore } from '../../../../../shared/stores/userStore';

type NavGroup = {
  label: string;
  items: { label: string; path: string; icon: React.ComponentType<{ size?: number }> }[];
};

function getNavGroups(isOwner: boolean): NavGroup[] {
  const groups: NavGroup[] = [
    {
      label: 'Account',
      items: [
        { label: 'Account', path: '/settings/account', icon: User },
        { label: 'AI Usage', path: '/settings/ai', icon: Sparkles },
        ...(isOwner
          ? [
              { label: 'General', path: '/settings/general', icon: Building2 },
              { label: 'Industry', path: '/settings/industry', icon: Briefcase },
            ]
          : []),
      ],
    },
  ];

  if (isOwner) {
    groups.push({
      label: 'Online Presence',
      items: [
        { label: 'Custom Domain', path: '/settings/domain', icon: Globe },
        { label: 'Email', path: '/settings/email', icon: Mail },
      ],
    });

    groups.push({
      label: 'Plans & Addons',
      items: [
        { label: 'Billing', path: '/settings/billing', icon: CreditCard },
        { label: 'Addons', path: '/settings/addons', icon: Puzzle },
        { label: 'Integrations', path: '/settings/integrations', icon: Layers },
      ],
    });
  }

  return groups;
}

const SettingsLayout: React.FC = () => {
  const navigate = useNavigate();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const { user } = useUserStore();
  const isOwner = user?.isOwner !== false;
  const navGroups = getNavGroups(isOwner);

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--dash-bg)' }}>
      {/* Sidebar */}
      <aside
        className="w-56 flex flex-col shrink-0 overflow-y-auto"
        style={{ background: 'var(--dash-surface)', borderRight: '1px solid var(--dash-border)' }}
      >
        {/* Header */}
        <div
          className="flex items-center gap-2 px-4 py-3.5 shrink-0"
          style={{ borderBottom: '1px solid var(--dash-border)' }}
        >
          <button
            onClick={() => navigate({ to: '/' })}
            className="h-6 w-6 rounded-md flex items-center justify-center transition-colors hover:bg-[var(--dash-hover)]"
            style={{ color: 'var(--dash-text-muted)' }}
          >
            <ArrowLeft size={14} />
          </button>
          <span className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
            Settings
          </span>
        </div>

        {/* Nav groups */}
        <nav className="flex-1 py-3 px-2 space-y-4">
          {navGroups.map((group) => (
            <div key={group.label}>
              <p
                className="px-3 mb-1 text-[10px] font-semibold uppercase tracking-widest"
                style={{ color: 'var(--dash-text-dim)' }}
              >
                {group.label}
              </p>
              <div className="space-y-0.5">
                {group.items.map(({ label, path, icon: Icon }) => {
                  const isActive = pathname === path;
                  return (
                    <Link
                      key={path}
                      to={path}
                      className={cn(
                        'flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors',
                        isActive ? 'bg-[var(--dash-active)]' : 'hover:bg-[var(--dash-hover)]'
                      )}
                      style={{
                        color: isActive ? 'var(--dash-text)' : 'var(--dash-text-muted)',
                        fontWeight: isActive ? 500 : undefined,
                      }}
                    >
                      <Icon size={14} {...({ className: 'shrink-0' } as any)} />
                      {label}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>
      </aside>

      {/* Content */}
      <main className="flex-1 overflow-y-auto p-8" style={{ color: 'var(--dash-text)' }}>
        <Outlet />
      </main>
    </div>
  );
};

export default SettingsLayout;
