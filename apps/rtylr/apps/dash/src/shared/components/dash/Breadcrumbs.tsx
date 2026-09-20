import React from 'react';
import { Link, useRouterState } from '@tanstack/react-router';
import { ChevronRight, Home } from 'lucide-react';
import { useTranslation } from 'react-i18next';

const ROUTE_KEYS: Record<string, string> = {
  '': 'hub',
  team: 'team',
  invite: 'addMember',
  permissions: 'permissions',
  roles: 'roles',
  apps: 'apps',
  activity: 'activityLog',
  notifications: 'notifications',
  account: 'account',
  billing: 'billing',
  integrations: 'integrations',
  'api-keys': 'apiKeys',
  webhooks: 'webhooks',
  audit: 'auditLog',
  security: 'security',
  settings: 'settings',
  subscription: 'billing',
  menus: 'menus',
  menu: 'menu',
  orders: 'orders',
  'qr-codes': 'qrCodes',
  design: 'design',
  analytics: 'analytics',
  marketplace: 'marketplace',
  changelog: 'whatsNew',
  flags: 'featureFlags',
  help: 'helpCenter',
  referrals: 'referrals',
  export: 'exportData',
  import: 'importData',
  theme: 'theme',
};

export const Breadcrumbs: React.FC<{ className?: string }> = ({ className }) => {
  const { t } = useTranslation('common');
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const segments = pathname.split('/').filter(Boolean);
  if (segments.length === 0) return null;

  return (
    <nav className={className} aria-label={t('common.breadcrumb')}>
      <ol className="flex items-center gap-1 text-xs" style={{ color: 'var(--dash-text-dim)' }}>
        <li>
          <Link to="/" className="flex items-center gap-1 hover:opacity-70 transition-opacity">
            <Home size={11} />
            {t('common.nav.hub', 'Hub')}
          </Link>
        </li>
        {segments.map((seg, i) => {
          const path = '/' + segments.slice(0, i + 1).join('/');
          const isLast = i === segments.length - 1;
          const key = ROUTE_KEYS[seg];
          const label = key ? t(`common.nav.${key}`) : seg;
          return (
            <React.Fragment key={path}>
              <ChevronRight size={11} className="shrink-0 rtl:rotate-180" />
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
