import { Outlet } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { APP_URLS } from '@voxire/config';
import { useUserStore } from '../stores/userStore';
import api from '../lib/api';

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const getAuthUrl = (): string => `${APP_URLS.auth}?redirect=hub`;

const APP_META: Record<string, { name: string; color: string; href: string }> = {
  pos:     { name: 'POS',     color: '#4583C4', href: APP_URLS.pos || 'https://pos.rtylr.com' },
  erp:     { name: 'ERP',     color: '#10b981', href: APP_URLS.erp || 'https://erp.rtylr.com' },
  hr:      { name: 'HR',      color: '#f59e0b', href: APP_URLS.hr || 'https://hr.rtylr.com' },
  crm:     { name: 'CRM',     color: '#8b5cf6', href: APP_URLS.crm || 'https://crm.rtylr.com' },
  finance: { name: 'Finance', color: '#ef4444', href: APP_URLS.finance || 'https://finance.rtylr.com' },
  flow:    { name: 'Flow',    color: '#64C5B9', href: APP_URLS.flow || 'https://flow.rtylr.com' },
};

function StaffView({ user }: { user: any }) {
  const apps: string[] = user?.apps ?? [];
  const businessName = user?.businessName || '';
  const fullName = user?.fullName || user?.email || '';

  const handleLogout = () => {
    document.cookie = 'token=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
    window.location.href = getAuthUrl();
  };

  return (
    <div style={{ minHeight: '100vh', background: '#fafafa', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '40px 20px' }}>
      <div style={{ maxWidth: 560, width: '100%', textAlign: 'center' }}>
        <img src="/logo.png" alt="rtylr" style={{ height: 28, marginBottom: 32 }} />
        {businessName && (
          <p style={{ fontSize: 13, color: '#999', marginBottom: 4 }}>{businessName}</p>
        )}
        <h1 style={{ fontSize: 22, fontWeight: 600, color: '#111', margin: '0 0 8px 0' }}>
          Welcome, {fullName.split(' ')[0]}
        </h1>
        <p style={{ fontSize: 14, color: '#666', margin: '0 0 32px 0' }}>
          Choose an app to get started.
        </p>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
          {apps.map((appId) => {
            const meta = APP_META[appId];
            if (!meta) return null;
            return (
              <a
                key={appId}
                href={meta.href}
                style={{
                  display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 10,
                  padding: '24px 16px', borderRadius: 12, background: '#fff',
                  border: '1px solid #e5e5e5', textDecoration: 'none',
                  transition: 'border-color 0.2s, box-shadow 0.2s',
                }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = meta.color; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = '#e5e5e5'; }}
              >
                <div style={{
                  width: 40, height: 40, borderRadius: 10,
                  background: `${meta.color}15`, display: 'flex',
                  alignItems: 'center', justifyContent: 'center',
                  fontSize: 12, fontWeight: 700, color: meta.color,
                }}>
                  {meta.name.charAt(0)}
                </div>
                <span style={{ fontSize: 13, fontWeight: 600, color: '#111' }}>{meta.name}</span>
              </a>
            );
          })}
        </div>

        <button
          onClick={handleLogout}
          style={{
            marginTop: 32, fontSize: 13, color: '#999', background: 'none',
            border: 'none', cursor: 'pointer', padding: '8px 16px',
          }}
        >
          Sign out
        </button>
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

    api.get('/user/me')
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
      <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <p style={{ color: '#999', fontSize: 14 }}>Loading...</p>
      </div>
    );
  }

  // Staff users (non-owner) get a simple app-cards view
  if (user && user.isOwner === false) {
    return <StaffView user={user} />;
  }

  // Owner/admin gets the full hub with sidebar
  return <Outlet />;
}

export { BusinessRoute };
