import { Outlet, useRouterState } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { APP_URLS, getAvatarUrl, resolveAuthUrl } from '@voxire/config';
import { signOut } from '@voxire/auth';
import { parseEffectiveAccess } from '@voxire/auth';
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
import { Avatar } from '@voxire/ui';
import { useUserStore, type StoredUser } from '../stores/userStore';
import { normalizeConfigData, useConfigStore } from '../stores/configStore';
import { accessClient, configClient, userClient } from '../lib/clients';
import { useBusinessFavicon } from '../hooks/useBusinessFavicon';

const getCookieToken = (): string | null => {
  const raw = document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.slice('token='.length);
  if (!raw) return null;
  try {
    return decodeURIComponent(raw);
  } catch {
    return raw;
  }
};

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

// Route path prefixes that billing/menu_manager may reach in addition to
// staffAllowedPaths, derived from their canonical resource grants in
// packages/permissions/src/matrix.ts (billing: 'hub:billing'; menu_manager:
// 'app:menu' + 'hub:design') and their invite descriptions in InviteUser.tsx.
// This is a route-level approximation of that matrix - Dash's router does not
// yet check resources per-route - so keep it narrow: it must not grow to
// include Team management, business deletion, or any other settings surface,
// which stay admin/owner (and operator, see hasManagementRole below) only.
const ROLE_SCOPED_PATHS: Record<string, string[]> = {
  billing: ['/settings/billing'],
  menu_manager: ['/menu', '/create-item', '/create-banner', '/qr-codes', '/design'],
};

function StaffView({ user }: { user: StoredUser }): React.JSX.Element {
  const userApps: string[] =
    user.effectiveAccess?.apps
      .filter((app) => app.licensed && app.hasAccess)
      .map((app) => app.appId) ??
    user.apps ??
    [];
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
      await accessClient.requestAccess({ app: appId });
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
        <img src="/logo-dark.png" alt="rtylr" style={{ height: 36, width: 'auto' }} />
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

function MFAAccessGate({ enrollmentRequired }: { enrollmentRequired: boolean }): React.JSX.Element {
  const { t } = useTranslation('common');

  return (
    <div
      className="flex min-h-screen items-center justify-center p-6"
      style={{ background: 'var(--dash-bg)' }}
    >
      <main
        className="w-full max-w-md rounded-2xl border p-6 text-center shadow-sm"
        style={{ background: 'var(--dash-surface)', borderColor: 'var(--dash-border)' }}
      >
        <h1 className="text-xl font-semibold" style={{ color: 'var(--dash-text)' }}>
          {enrollmentRequired ? t('mfaGate.enrollmentTitle') : t('mfaGate.verificationTitle')}
        </h1>
        <p className="mt-2 text-sm" style={{ color: 'var(--dash-text-muted)' }}>
          {enrollmentRequired
            ? t('mfaGate.enrollmentDescription')
            : t('mfaGate.verificationDescription')}
        </p>
        <a
          href="/security"
          className="mt-5 inline-flex min-h-11 items-center justify-center rounded-lg px-4 text-sm font-semibold"
          style={{ background: 'var(--accent)', color: 'var(--accent-contrast, #fff)' }}
        >
          {enrollmentRequired ? t('mfaGate.openSetup') : t('mfaGate.openVerification')}
        </a>
      </main>
    </div>
  );
}

function BusinessRoute(): React.JSX.Element | null {
  const token = getCookieToken();
  const { user, setUser, setToken } = useUserStore();
  const { setWebsiteData, setContactInfo, setMaintanance, setAboutContent } = useConfigStore();
  const logoUrl = useConfigStore((s) => s.websiteData?.logoUrl as string | undefined);
  const [loading, setLoading] = useState(true);
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  useBusinessFavicon(logoUrl);

  useEffect((): void => {
    if (!token) {
      window.location.href = `${resolveAuthUrl()}?redirect=dash`;
      return;
    }

    setToken(token);

    userClient
      .getMe({})
      .then((res) => {
        const effectiveAccess = parseEffectiveAccess(res.effectiveAccess);
        setUser({
          ...res,
          dashRole: res.dashRole ?? effectiveAccess?.dashRole ?? undefined,
          effectiveAccess: effectiveAccess ?? undefined,
          provisioningStatus: res.provisioningStatus
            ? {
                mode: res.provisioningStatus.mode ?? 'invitation',
                state: res.provisioningStatus.state ?? 'pending_activation',
                emailVerified: res.provisioningStatus.emailVerified ?? false,
                passwordChangeRequired: res.provisioningStatus.passwordChangeRequired ?? false,
                mfaRequired: res.provisioningStatus.mfaRequired ?? false,
                mfaEnrolled: res.provisioningStatus.mfaEnrolled ?? false,
                mfaVerified: res.provisioningStatus.mfaVerified ?? false,
                recoveryAvailable: res.provisioningStatus.recoveryAvailable ?? false,
              }
            : undefined,
        } as StoredUser);
        configClient
          .getConfig({})
          .then((configRes) => {
            const business = configRes.business ?? {};
            const config = normalizeConfigData({
              site_name: business.name,
              logo_url: business.logoUrl,
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
            setMaintanance(configRes.maintenance ?? false);
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
  // businessId arrives over the wire as an int64-string (e.g. "0"), so it must
  // be parsed with Number() before a truthiness check - a bare cast leaves it
  // a non-empty string, and `!"0"` is false.
  const currentBusinessId = Number(user?.businessId ?? 0);
  if (user && !currentBusinessId) {
    return <CreateBusinessOnboarding />;
  }

  const provisioningStatus = user?.provisioningStatus;
  if (
    user &&
    provisioningStatus?.mfaRequired &&
    !provisioningStatus.mfaVerified &&
    pathname !== '/security'
  ) {
    return <MFAAccessGate enrollmentRequired={!provisioningStatus.mfaEnrolled} />;
  }

  // The normalized access document is authoritative when present. Legacy role
  // routing remains only for the migration window when the server did not send
  // that document.
  const role = (user?.role ?? '').toLowerCase();
  const hasManagementRole = user?.effectiveAccess
    ? user.effectiveAccess.dashPermissions.some(
        (permission) =>
          permission === 'hub:team:edit' ||
          permission === 'hub:roles:view' ||
          permission === 'hub:licenses:manage'
      )
    : role === 'admin';
  const staffAllowedPaths = user?.effectiveAccess
    ? ['/settings/account']
    : ['/settings/account', ...(ROLE_SCOPED_PATHS[role] ?? [])];

  if (
    user &&
    user.isOwner === false &&
    !hasManagementRole &&
    !staffAllowedPaths.some((p) => pathname.startsWith(p))
  ) {
    return <StaffView user={user} />;
  }

  // Owners, admins/operators, and role-scoped members (on their allowed
  // paths above) reach here.
  return <Outlet />;
}

export { BusinessRoute };
