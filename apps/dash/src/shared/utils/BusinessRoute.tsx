import { Outlet, useRouterState } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { APP_URLS, getAvatarUrl, resolveAuthUrl } from '@voxire/config';
import { signOut } from '@voxire/auth';
import CreateBusinessOnboarding from '../../features/onboarding/CreateBusinessOnboarding';
import {
  MonitorSpeaker,
  Package,
  UserCheck,
  Users,
  CreditCard,
  Layers,
  Lock,
  BarChart3,
} from 'lucide-react';
import { unwrap } from '@voxire/api';
import { Avatar } from '@voxire/ui';
import { useUserStore, type StoredUser } from '../stores/userStore';
import { normalizeConfigData, useConfigStore } from '../stores/configStore';
import { safe } from '../lib/api';

const getCookieToken = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const ALL_APPS = [
  {
    id: 'pos',
    name: 'POS',
    desc: 'Point of sale',
    color: 'var(--app-pos)',
    icon: MonitorSpeaker,
    href: APP_URLS.pos,
  },
  {
    id: 'erp',
    name: 'ERP',
    desc: 'Inventory and supply',
    color: 'var(--app-erp)',
    icon: Package,
    href: APP_URLS.erp,
  },
  {
    id: 'hr',
    name: 'HR',
    desc: 'Staff and payroll',
    color: 'var(--app-hr)',
    icon: UserCheck,
    href: APP_URLS.hr,
  },
  {
    id: 'crm',
    name: 'CRM',
    desc: 'Customers and loyalty',
    color: 'var(--app-crm)',
    icon: Users,
    href: APP_URLS.crm,
  },
  {
    id: 'finance',
    name: 'Finance',
    desc: 'Invoices and expenses',
    color: 'var(--app-finance)',
    icon: CreditCard,
    href: APP_URLS.finance,
  },
  {
    id: 'flow',
    name: 'Flow',
    desc: 'Projects and tasks',
    color: 'var(--app-flow)',
    icon: Layers,
    href: APP_URLS.flow,
  },
  {
    id: 'insights',
    name: 'Insights',
    desc: 'Analytics and reports',
    color: 'var(--app-insights)',
    icon: BarChart3,
    href: APP_URLS.insights,
  },
];

function StaffView({ user }: { user: StoredUser }): React.JSX.Element {
  const userApps: string[] = user?.apps ?? [];
  const businessName = user?.businessName || '';
  const fullName = user?.fullName || user?.email || '';
  const [requested, setRequested] = useState<Set<string>>(new Set());
  const [requestLoading, setRequestLoading] = useState<string | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);

  const handleLogout = (): void => {
    signOut('dash');
  };

  const handleRequestAccess = async (appId: string): Promise<void> => {
    setRequestLoading(appId);
    try {
      await safe.post('/hub/team/access/v1/requests', { app: appId });
      setRequested((prev) => new Set(prev).add(appId));
    } catch {
      // silently fail, server returns success anyway if already requested
    } finally {
      setRequestLoading(null);
    }
  };

  const avatarUrl = getAvatarUrl(user?.email || fullName);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--dash-bg, #fafafa)' }}>
      {/* Top bar */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '16px 32px',
          borderBottom: '1px solid #eee',
          background: 'var(--dash-bg, #fff)',
        }}
      >
        <img src="/logo.png" alt="rtylr" style={{ height: 24, filter: 'brightness(0)' }} />
        <div style={{ position: 'relative' }}>
          <button
            type="button"
            onClick={() => setMenuOpen(!menuOpen)}
            aria-label="User menu"
            style={{ border: 'none', padding: 0, background: 'transparent', cursor: 'pointer' }}
          >
            <Avatar src={avatarUrl} alt={fullName || 'User'} size="md" />
          </button>
          {menuOpen && (
            <div
              style={{
                position: 'absolute',
                top: 44,
                right: 0,
                width: 200,
                background: 'var(--dash-bg, #fff)',
                borderRadius: 10,
                border: '1px solid var(--dash-border, #e5e5e5)',
                boxShadow: '0 4px 16px rgba(0,0,0,0.08)',
                zIndex: 50,
                overflow: 'hidden',
              }}
            >
              <div style={{ padding: '12px 16px', borderBottom: '1px solid #f0f0f0' }}>
                <p
                  style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: 'var(--dash-text, #111)',
                    margin: 0,
                  }}
                >
                  {fullName}
                </p>
                {businessName && (
                  <p
                    style={{
                      fontSize: 11,
                      color: 'var(--dash-text-muted, #999)',
                      margin: '2px 0 0 0',
                    }}
                  >
                    {businessName}
                  </p>
                )}
              </div>
              <a
                href="/settings/account"
                style={{
                  display: 'block',
                  padding: '10px 16px',
                  fontSize: 13,
                  color: 'var(--dash-text, #333)',
                  textDecoration: 'none',
                }}
              >
                Account settings
              </a>
              <button
                type="button"
                onClick={handleLogout}
                style={{
                  display: 'block',
                  width: '100%',
                  padding: '10px 16px',
                  fontSize: 13,
                  color: '#ef4444',
                  background: 'none',
                  border: 'none',
                  textAlign: 'left',
                  cursor: 'pointer',
                }}
              >
                Sign out
              </button>
            </div>
          )}
        </div>
      </div>

      {/* App grid */}
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          padding: '48px 32px',
        }}
      >
        <div style={{ maxWidth: 960, width: '100%', textAlign: 'center' }}>
          <h1
            style={{
              fontSize: 24,
              fontWeight: 600,
              color: 'var(--dash-text, #111)',
              margin: '0 0 6px 0',
            }}
          >
            Welcome, {fullName.split(' ')[0]}
          </h1>
          {businessName && (
            <p style={{ fontSize: 13, color: 'var(--dash-text-muted, #999)', marginBottom: 0 }}>
              {businessName}
            </p>
          )}
          <div style={{ height: 36 }} />

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: 12 }}>
            {ALL_APPS.map((app) => {
              const Icon = app.icon;
              const hasAccess = userApps.includes(app.id);
              const isRequested = requested.has(app.id);

              if (hasAccess) {
                return (
                  <a
                    key={app.id}
                    href={app.href}
                    style={{
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: 8,
                      padding: '24px 16px',
                      borderRadius: 12,
                      background: 'var(--dash-bg, #fff)',
                      border: '1px solid var(--dash-border, #e5e5e5)',
                      textDecoration: 'none',
                      transition: 'border-color 0.2s',
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.borderColor = app.color;
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.borderColor = '#e5e5e5';
                    }}
                  >
                    <div
                      style={{
                        width: 40,
                        height: 40,
                        borderRadius: 10,
                        background: `${app.color}12`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      <Icon size={20} color={app.color} />
                    </div>
                    <span
                      style={{ fontSize: 13, fontWeight: 600, color: 'var(--dash-text, #111)' }}
                    >
                      {app.name}
                    </span>
                    <span style={{ fontSize: 11, color: 'var(--dash-text-muted, #999)' }}>
                      {app.desc}
                    </span>
                  </a>
                );
              }

              return (
                <div
                  key={app.id}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: 8,
                    padding: '24px 16px',
                    borderRadius: 12,
                    background: 'var(--dash-bg, #fafafa)',
                    border: '1px dashed var(--dash-border, #e0e0e0)',
                    opacity: 0.7,
                  }}
                >
                  <div
                    style={{
                      width: 40,
                      height: 40,
                      borderRadius: 10,
                      background: 'var(--dash-hover, #f0f0f0)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <Lock size={16} color="#bbb" />
                  </div>
                  <span
                    style={{ fontSize: 13, fontWeight: 600, color: 'var(--dash-text-muted, #888)' }}
                  >
                    {app.name}
                  </span>
                  {isRequested ? (
                    <span style={{ fontSize: 11, color: '#10b981', fontWeight: 500 }}>
                      Requested
                    </span>
                  ) : requestLoading === app.id ? (
                    <div
                      style={{
                        width: 14,
                        height: 14,
                        border: '2px solid #e5e5e5',
                        borderTopColor: '#4583C4',
                        borderRadius: '50%',
                        animation: 'spin 0.6s linear infinite',
                      }}
                    />
                  ) : (
                    <button
                      type="button"
                      onClick={() => handleRequestAccess(app.id)}
                      style={{
                        fontSize: 11,
                        color: '#4583C4',
                        background: 'none',
                        border: 'none',
                        cursor: 'pointer',
                        fontWeight: 500,
                        padding: 0,
                        textDecoration: 'underline',
                        textUnderlineOffset: 2,
                      }}
                    >
                      Request access
                    </button>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}

function BusinessRoute(): React.JSX.Element | null {
  const token = getCookieToken();
  const { user, setUser, setToken } = useUserStore();
  const { setWebsiteData, setContactInfo, setMaintanance, setAboutContent } = useConfigStore();
  const [loading, setLoading] = useState(true);
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  useEffect((): void => {
    if (!token) {
      window.location.href = `${resolveAuthUrl()}?redirect=dash`;
      return;
    }

    setToken(token);

    unwrap(safe.get<Record<string, unknown>>('/hub/user/v1/users/me'))
      .then((res) => {
        setUser(res as StoredUser);
        unwrap(
          safe.get<{ business?: Record<string, unknown>; maintenance?: boolean }>(
            '/hub/config/v1/config'
          )
        )
          .then((configRes) => {
            const business = (configRes?.business ?? {}) as Record<string, unknown>;
            const config = normalizeConfigData({
              site_name: business.name,
              logo_url: business.logo_url ?? business.logoUrl,
              contact_email: business.email,
              contact_number: business.phone,
              address: business.address,
              slug: business.slug,
              website: business.website,
              country: business.country,
            });
            setWebsiteData(config);
            setContactInfo({
              contactEmail: (config.contactEmail as string | null) ?? null,
              contactNumber: (config.contactNumber as string | null) ?? null,
            });
            setMaintanance((configRes?.maintenance as boolean) ?? false);
            setAboutContent({ aboutContent: config.aboutContent ?? null });
          })
          .catch(() => {
            // config is non-critical - missing config won't block the app
          });
      })
      .catch(() => {
        // auth interceptor handles 401 redirect
      })
      .finally(() => setLoading(false));
  }, [token]);

  if (!token) return null;

  if (loading) {
    return (
      <div
        style={{
          height: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'var(--dash-bg, #fff)',
        }}
      >
        <div
          style={{
            width: 24,
            height: 24,
            border: '2px solid #e5e5e5',
            borderTopColor: '#4583C4',
            borderRadius: '50%',
            animation: 'spin 0.6s linear infinite',
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
      </div>
    );
  }

  // Freshly-registered owners have no business yet (registration no longer
  // auto-creates one). Gate them into onboarding to create their first business
  // before anything else loads.
  const currentBusinessId = (user?.businessId as number | undefined) ?? 0;
  if (user && !currentBusinessId) {
    return <CreateBusinessOnboarding />;
  }

  // Staff users (non-owner) see the app cards launcher.
  // Account settings are always accessible regardless of role.
  const staffAllowedPaths = ['/settings/account'];
  if (user && user.isOwner === false && !staffAllowedPaths.some((p) => pathname.startsWith(p))) {
    return <StaffView user={user} />;
  }

  // Owners and admins get the full Dash experience.
  return <Outlet />;
}

export { BusinessRoute };
