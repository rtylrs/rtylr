import React, { useState } from 'react';
import { Link, Outlet, useLocation } from '@tanstack/react-router';
import { Briefcase, Inbox, Palette, LayoutGrid, Menu, X, LogOut } from 'lucide-react';
import { signOut, useUserStore } from '@voxire/auth';
import { canEffectiveAccess } from '@voxire/permissions';

/**
 * AdminLayout - wraps the authenticated admin routes under /admin.
 * Responsive layout supporting mobile, tablet, and desktop viewports
 * without horizontal overflow. Meets the 44px touch target minimum.
 */
const AdminLayout: React.FC = (): React.JSX.Element => {
  const user = useUserStore((s) => s.user);
  const effectiveAccess = user?.effectiveAccess;
  const canRecruit = (permission: string): boolean =>
    effectiveAccess ? canEffectiveAccess(effectiveAccess, permission) : false;
  const canAdministerRecruit =
    canRecruit('recruit:jobs:view') ||
    canRecruit('recruit:jobs:manage') ||
    canRecruit('recruit:candidates:view');
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  if (user && !canAdministerRecruit) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-[var(--recruit-admin-bg)] p-6 text-center text-white/70">
        <h2 className="text-xl font-semibold text-white">Access Denied</h2>
        <p className="mt-2 text-sm text-white/60">
          Owner or Admin permission is required to access Recruit administration and candidate data.
        </p>
      </div>
    );
  }

  const navItems = [
    {
      to: '/admin/careers',
      icon: <Briefcase className="h-4 w-4" />,
      label: 'Careers',
      visible: canRecruit('recruit:jobs:view') || canRecruit('recruit:jobs:manage'),
    },
    {
      to: '/admin/submissions',
      icon: <Inbox className="h-4 w-4" />,
      label: 'Submissions',
      visible: canRecruit('recruit:candidates:view'),
    },
    {
      to: '/admin/sections',
      icon: <LayoutGrid className="h-4 w-4" />,
      label: 'Page sections',
      visible: canRecruit('recruit:jobs:manage'),
    },
    {
      to: '/admin/customization',
      icon: <Palette className="h-4 w-4" />,
      label: 'Page branding',
      visible: canRecruit('recruit:jobs:manage'),
    },
  ].filter((item) => item.visible);

  return (
    <div
      className="min-h-screen"
      style={{ background: 'var(--recruit-admin-bg)', color: 'var(--recruit-admin-fg)' }}
    >
      {/* Mobile Top Bar */}
      <div
        className="flex items-center justify-between border-b px-4 py-3 md:hidden"
        style={{
          borderColor: 'var(--recruit-admin-border)',
          background: 'var(--recruit-admin-surface)',
        }}
      >
        <div className="text-base font-semibold">Recruit Admin</div>
        <button
          type="button"
          onClick={() => setMobileNavOpen((prev) => !prev)}
          className="flex h-11 w-11 items-center justify-center rounded-md border border-white/10 text-white/80 hover:bg-white/5"
          aria-label="Toggle navigation menu"
        >
          {mobileNavOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {/* Mobile Dropdown Nav */}
      {mobileNavOpen && (
        <div
          className="border-b px-3 py-3 md:hidden"
          style={{
            borderColor: 'var(--recruit-admin-border)',
            background: 'var(--recruit-admin-surface)',
          }}
        >
          <nav className="space-y-1">
            {navItems.map((item) => (
              <AdminLink
                key={item.to}
                to={item.to}
                icon={item.icon}
                label={item.label}
                onClick={() => setMobileNavOpen(false)}
              />
            ))}
          </nav>
          <div
            className="mt-3 border-t pt-3"
            style={{ borderColor: 'var(--recruit-admin-border)' }}
          >
            <button
              type="button"
              onClick={() => signOut('recruit')}
              className="flex min-h-[44px] w-full items-center gap-2 rounded-md px-3 py-2.5 text-sm font-medium text-red-400 hover:bg-white/5"
            >
              <LogOut className="h-4 w-4" />
              Sign out
            </button>
          </div>
        </div>
      )}

      {/* Desktop Sidebar */}
      <aside
        className="hidden md:fixed md:left-0 md:top-0 md:flex md:h-screen md:w-56 md:flex-col md:border-r"
        style={{
          borderColor: 'var(--recruit-admin-border)',
          background: 'var(--recruit-admin-surface)',
        }}
      >
        <div className="px-5 py-5 text-lg font-semibold">Recruit</div>
        <nav className="flex-1 space-y-1 px-3">
          {navItems.map((item) => (
            <AdminLink key={item.to} to={item.to} icon={item.icon} label={item.label} />
          ))}
        </nav>
        <div className="mx-3 mb-4 flex flex-col gap-1">
          <button
            type="button"
            onClick={() => signOut('recruit')}
            className="flex min-h-[44px] items-center gap-2 rounded-md border-0 bg-transparent px-3 py-2.5 text-left text-sm font-medium text-red-400 cursor-pointer hover:bg-white/5 hover:text-red-300"
          >
            <LogOut className="h-4 w-4" />
            Sign out
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <main className="min-h-screen w-full min-w-0 overflow-x-hidden md:ml-56">
        <Outlet />
      </main>
    </div>
  );
};

const AdminLink: React.FC<{
  to: string;
  icon: React.ReactNode;
  label: string;
  onClick?: () => void;
}> = ({ to, icon, label, onClick }) => {
  const location = useLocation();
  const isActive = location.pathname === to || location.pathname.startsWith(to + '/');
  return (
    <Link
      to={to as any}
      onClick={onClick}
      className={`flex min-h-[44px] items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition ${
        isActive ? 'bg-white/10 text-white' : 'text-white/70 hover:bg-white/5 hover:text-white'
      }`}
    >
      {icon}
      {label}
    </Link>
  );
};

export default AdminLayout;
