import React from 'react';
import { Link, useRouterState } from '@tanstack/react-router';
import { ChevronRight, Home } from 'lucide-react';

const ROUTE_LABELS: Record<string, string> = {
  '': 'Hub',
  team: 'Team',
  invite: 'Invite',
  permissions: 'Permissions',
  roles: 'Roles',
  apps: 'Apps',
  activity: 'Activity',
  notifications: 'Notifications',
  account: 'Account',
  billing: 'Billing',
  integrations: 'Integrations',
  'api-keys': 'API Keys',
  webhooks: 'Webhooks',
  audit: 'Audit',
  security: 'Security',
  settings: 'Settings',
  subscription: 'Billing',
  menus: 'Menus',
  menu: 'Menu',
  orders: 'Orders',
  'qr-codes': 'QR Codes',
  design: 'Design',
  analytics: 'Analytics',
  marketplace: 'Marketplace',
  changelog: "What's new",
  flags: 'Feature Flags',
  help: 'Help Center',
  referrals: 'Referrals',
  export: 'Export Data',
  import: 'Import Data',
  theme: 'Theme',
};

export const Breadcrumbs: React.FC<{ className?: string }> = ({ className }) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const segments = pathname.split('/').filter(Boolean);
  if (segments.length === 0) return null;

  return (
    <nav className={className} aria-label="Breadcrumb">
      <ol className="flex items-center gap-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
        <li>
          <Link to="/" className="flex items-center gap-1 hover:opacity-70 transition-opacity">
            <Home size={11} />
            Hub
          </Link>
        </li>
        {segments.map((seg, i) => {
          const path = '/' + segments.slice(0, i + 1).join('/');
          const isLast = i === segments.length - 1;
          const label = ROUTE_LABELS[seg] ?? seg;
          return (
            <React.Fragment key={path}>
              <ChevronRight size={11} className="shrink-0" />
              <li>
                {isLast ? (
                  <span style={{ color: 'var(--dash-text)' }} className="font-medium">
                    {label}
                  </span>
                ) : (
                  <Link to={path} className="hover:opacity-70 transition-opacity">
                    {label}
                  </Link>
                )}
              </li>
            </React.Fragment>
          );
        })}
      </ol>
    </nav>
  );
};

export default Breadcrumbs;
