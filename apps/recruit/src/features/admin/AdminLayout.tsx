import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { Briefcase, Inbox, Palette, ArrowLeft, LayoutGrid } from 'lucide-react';
import { APP_URLS } from '@voxire/config';

/**
 * AdminLayout — wraps the authenticated admin routes under /admin.
 * The public careers pages share the same app but render through the
 * template resolver and bypass this shell entirely.
 */
const AdminLayout: React.FC = (): React.JSX.Element => {
  return (
    <div className="min-h-screen" style={{ background: '#0a0a0a', color: '#fafaf9' }}>
      <aside
        className="fixed left-0 top-0 flex h-screen w-56 flex-col border-r"
        style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
      >
        <div className="px-5 py-5 text-lg font-semibold">Recruit</div>
        <nav className="flex-1 space-y-1 px-3">
          <AdminLink to="/admin/careers" icon={<Briefcase className="h-4 w-4" />} label="Careers" />
          <AdminLink
            to="/admin/submissions"
            icon={<Inbox className="h-4 w-4" />}
            label="Submissions"
          />
          <AdminLink
            to="/admin/sections"
            icon={<LayoutGrid className="h-4 w-4" />}
            label="Page sections"
          />
          <AdminLink
            to="/admin/customization"
            icon={<Palette className="h-4 w-4" />}
            label="Page branding"
          />
        </nav>
        <a
          href={APP_URLS.dash}
          className="mx-3 mb-4 inline-flex items-center gap-2 rounded-md px-3 py-2 text-xs"
          style={{ color: '#999' }}
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to dashboard
        </a>
      </aside>
      <main className="ml-56 min-h-screen">
        <Outlet />
      </main>
    </div>
  );
};

const AdminLink: React.FC<{ to: string; icon: React.ReactNode; label: string }> = ({
  to,
  icon,
  label,
}) => (
  <NavLink
    to={to}
    className={({ isActive }) =>
      `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition ${
        isActive ? 'bg-white/5 text-white' : 'text-white/70 hover:bg-white/5 hover:text-white'
      }`
    }
  >
    {icon}
    {label}
  </NavLink>
);

export default AdminLayout;
