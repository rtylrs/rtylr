import React from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import {
  User, Building2, Globe, Mail, CreditCard, Puzzle,
  ArrowLeft, Briefcase, Layers,
} from 'lucide-react';
import { cn } from '../../../../../lib/utils';
import { useUserStore } from '../../../../../stores/userStore';

type NavGroup = {
  label: string;
  items: { label: string; path: string; icon: React.ComponentType<{ size?: number }> }[];
};

function getNavGroups(isOwner: boolean): NavGroup[] {
  const groups: NavGroup[] = [
    {
      label: 'Account',
      items: [
        { label: 'Account',  path: '/settings/account',  icon: User },
        ...(isOwner ? [
          { label: 'General',  path: '/settings/general',  icon: Building2 },
          { label: 'Industry', path: '/settings/industry', icon: Briefcase },
        ] : []),
      ],
    },
  ];

  if (isOwner) {
    groups.push({
      label: 'Online Presence',
      items: [
        { label: 'Custom Domain', path: '/settings/domain', icon: Globe },
        { label: 'Email',         path: '/settings/email',  icon: Mail },
      ],
    });

    groups.push({
      label: 'Plans & Addons',
      items: [
        { label: 'Billing',      path: '/settings/billing',      icon: CreditCard },
        { label: 'Addons',       path: '/settings/addons',       icon: Puzzle },
        { label: 'Integrations', path: '/settings/integrations', icon: Layers },
      ],
    });
  }

  return groups;
}

const SettingsLayout: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useUserStore();
  const isOwner = user?.isOwner !== false;
  const navGroups = getNavGroups(isOwner);

  return (
    <div className="flex h-screen overflow-hidden" style={{ background: 'var(--hub-bg)' }}>
      {/* Sidebar */}
      <aside
        className="w-56 flex flex-col shrink-0 overflow-y-auto"
        style={{ background: 'var(--hub-surface)', borderRight: '1px solid var(--hub-border)' }}
      >
        {/* Header */}
        <div
          className="flex items-center gap-2 px-4 py-3.5 shrink-0"
          style={{ borderBottom: '1px solid var(--hub-border)' }}
        >
          <button
            onClick={() => navigate('/')}
            className="h-6 w-6 rounded-md flex items-center justify-center transition-colors hover:bg-[var(--hub-hover)]"
            style={{ color: 'var(--hub-text-muted)' }}
          >
            <ArrowLeft size={14} />
          </button>
          <span className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            Settings
          </span>
        </div>

        {/* Nav groups */}
        <nav className="flex-1 py-3 px-2 space-y-4">
          {navGroups.map((group) => (
            <div key={group.label}>
              <p
                className="px-3 mb-1 text-[10px] font-semibold uppercase tracking-widest"
                style={{ color: 'var(--hub-text-dim)' }}
              >
                {group.label}
              </p>
              <div className="space-y-0.5">
                {group.items.map(({ label, path, icon: Icon }) => (
                  <NavLink
                    key={path}
                    to={path}
                    className={({ isActive }) =>
                      cn(
                        'flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors',
                        isActive ? 'bg-[var(--hub-active)]' : 'hover:bg-[var(--hub-hover)]'
                      )
                    }
                    style={({ isActive }) => ({
                      color: isActive ? 'var(--hub-text)' : 'var(--hub-text-muted)',
                      fontWeight: isActive ? 500 : undefined,
                    })}
                  >
                    <Icon size={14} className="shrink-0" />
                    {label}
                  </NavLink>
                ))}
              </div>
            </div>
          ))}
        </nav>
      </aside>

      {/* Content */}
      <main className="flex-1 overflow-y-auto p-8" style={{ color: 'var(--hub-text)' }}>
        <Outlet />
      </main>
    </div>
  );
};

export default SettingsLayout;
