import { Outlet, useLocation } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { APP_URLS } from '@voxire/config';
import {
  MonitorSpeaker,
  Package,
  UserCheck,
  Users,
  CreditCard,
  Layers,
  Lock,
  Store,
  CalendarCheck,
  Headphones,
  BarChart3,
} from 'lucide-react';
import { useUserStore } from '../stores/userStore';
import api from '../lib/api';

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const getAuthUrl = (): string => `${APP_URLS.auth}?redirect=hub`;

const ALL_APPS = [
  {
    id: 'pos',
    name: 'POS',
    desc: 'Point of sale',
    color: '#4583C4',
    icon: MonitorSpeaker,
    href: APP_URLS.pos || 'https://pos.rtylr.com',
  },
  {
    id: 'erp',
    name: 'ERP',
    desc: 'Inventory and supply',
    color: '#10b981',
    icon: Package,
    href: APP_URLS.erp || 'https://erp.rtylr.com',
  },
  {
    id: 'hr',
    name: 'HR',
    desc: 'Staff and payroll',
    color: '#f59e0b',
    icon: UserCheck,
    href: APP_URLS.hr || 'https://hr.rtylr.com',
  },
  {
    id: 'crm',
    name: 'CRM',
    desc: 'Customers and loyalty',
    color: '#8b5cf6',
    icon: Users,
    href: APP_URLS.crm || 'https://crm.rtylr.com',
  },
  {
    id: 'finance',
    name: 'Finance',
    desc: 'Invoices and expenses',
    color: '#ef4444',
    icon: CreditCard,
    href: APP_URLS.finance || 'https://finance.rtylr.com',
  },
  {
    id: 'flow',
    name: 'Flow',
    desc: 'Projects and tasks',
    color: '#64C5B9',
    icon: Layers,
    href: APP_URLS.flow || 'https://flow.rtylr.com',
  },
  {
    id: 'store',
    name: 'Store',
    desc: 'Online store and orders',
    color: '#14b8a6',
    icon: Store,
    href: APP_URLS.store || 'https://store.rtylr.com',
  },
  {
    id: 'book',
    name: 'Book',
    desc: 'Appointments and scheduling',
    color: '#0ea5e9',
    icon: CalendarCheck,
    href: APP_URLS.book || 'https://book.rtylr.com',
  },
  {
    id: 'support',
    name: 'Support',
    desc: 'Help desk and tickets',
    color: '#4583C4',
    icon: Headphones,
    href: APP_URLS.support || 'https://support.rtylr.com',
  },
  {
    id: 'insights',
    name: 'Insights',
    desc: 'Analytics and reports',
    color: '#6366f1',
    icon: BarChart3,
    href: APP_URLS.insights || 'https://insights.rtylr.com',
  },
];

function StaffView({ user }: { user: any }) {
  const userApps: string[] = user?.apps ?? [];
  const businessName = user?.businessName || '';
  const fullName = user?.fullName || user?.email || '';
  const [requested, setRequested] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState<string | null>(null);

  const handleLogout = () => {
    document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    window.location.href = getAuthUrl();
  };

  const handleRequestAccess = async (appId: string) => {
    setLoading(appId);
    try {
      await api.post('/team/request-access', { app: appId });
      setRequested((prev) => new Set(prev).add(appId));
    } catch {
      // Don't mark as requested if it failed
    } finally {
      setLoading(null);
    }
  };

  const initials = fullName
    .split(' ')
    .map((n: any) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--hub-bg, #fafafa)' }}>
      {/* Top bar */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '16px 32px',
          borderBottom: '1px solid #eee',
          background: 'var(--hub-bg, #fff)',
        }}
      >
        <img src="/logo.png" alt="rtylr" style={{ height: 24, filter: 'brightness(0)' }} />
        <div style={{ position: 'relative' }}>
          <button
            onClick={() => setMenuOpen(!menuOpen)}
            style={{
              width: 36,
              height: 36,
              borderRadius: '50%',
              border: 'none',
              cursor: 'pointer',
              background: '#4583C4',
              color: '#fff',
              fontSize: 13,
              fontWeight: 600,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            {initials}
          </button>
          {menuOpen && (
            <div
              style={{
                position: 'absolute',
                top: 44,
                right: 0,
                width: 200,
                background: 'var(--hub-bg, #fff)',
                borderRadius: 10,
                border: '1px solid var(--hub-border, #e5e5e5)',
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
                    color: 'var(--hub-text, #111)',
                    margin: 0,
                  }}
                >
                  {fullName}
                </p>
                {businessName && (
                  <p
                    style={{
                      fontSize: 11,
                      color: 'var(--hub-text-muted, #999)',
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
                  color: 'var(--hub-text, #333)',
                  textDecoration: 'none',
                }}
              >
                Account settings
              </a>
              <button
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

      {/* Main content */}
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
              color: 'var(--hub-text, #111)',
              margin: '0 0 6px 0',
            }}
          >
            Welcome, {fullName.split(' ')[0]}
          </h1>
          {businessName && (
            <p style={{ fontSize: 13, color: 'var(--hub-text-muted, #999)', marginBottom: 0 }}>
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
                      background: 'var(--hub-bg, #fff)',
                      border: '1px solid var(--hub-border, #e5e5e5)',
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
                    <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text, #111)' }}>
                      {app.name}
                    </span>
                    <span style={{ fontSize: 11, color: 'var(--hub-text-muted, #999)' }}>
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
                    background: 'var(--hub-bg, #fafafa)',
                    border: '1px dashed var(--hub-border, #e0e0e0)',
                    opacity: 0.7,
                  }}
                >
                  <div
                    style={{
                      width: 40,
                      height: 40,
                      borderRadius: 10,
                      background: 'var(--hub-hover, #f0f0f0)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <Lock size={16} color="#bbb" />
                  </div>
                  <span
                    style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text-muted, #888)' }}
                  >
                    {app.name}
                  </span>
                  {isRequested ? (
                    <span style={{ fontSize: 11, color: '#10b981', fontWeight: 500 }}>
                      Requested
                    </span>
                  ) : loading === app.id ? (
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

function BusinessRoute() {
  const token = getTokenFromCookie();
  const { user, setUser, setToken } = useUserStore();
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) {
      window.location.href = getAuthUrl();
      return;
    }

    setToken(token);

    api
      .get('/user/me')
      .then((res) => {
        if (res.data) {
          setUser(res.data);
        }
      })
      .catch(() => {})
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
          background: 'var(--hub-bg, #fff)',
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

  // Staff users (non-owner) get a simple app-cards view
  // But allow access to account settings
  const location = useLocation();
  const staffAllowedPaths = ['/settings/account'];
  if (
    user &&
    user.isOwner === false &&
    !staffAllowedPaths.some((p) => location.pathname.startsWith(p))
  ) {
    return <StaffView user={user} />;
  }

  // Owner/admin gets the full hub with sidebar
  return <Outlet />;
}

export { BusinessRoute };
