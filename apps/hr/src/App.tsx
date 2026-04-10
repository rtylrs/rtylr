import React, { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { APP_URLS } from '@voxire/config';
import AppRoutes from './routes/AppRoutes';

/** Read a cookie by name */
function getCookie(name: string): string | null {
  if (typeof document === 'undefined') return null;
  const nameEQ = `${name}=`;
  for (const part of document.cookie.split('; ')) {
    if (part.startsWith(nameEQ)) return decodeURIComponent(part.slice(nameEQ.length));
  }
  return null;
}

function AuthGuard({ children }: { children: React.ReactNode }) {
  const [allowed, setAllowed] = useState<boolean | null>(null);

  useEffect(() => {
    const token = getCookie('token');
    if (!token) {
      window.location.href = `${APP_URLS.auth}?redirect=hr`;
      return;
    }

    import('./lib/api').then(({ default: api }) => {
      api.get('/user/me')
        .then((res: any) => {
          const apps: string[] = res.data?.apps ?? [];
          const role: string = res.data?.role ?? '';
          if (role === 'owner' || role === 'admin' || apps.length === 0 || apps.includes('hr')) {
            setAllowed(true);
          } else {
            setAllowed(false);
          }
        })
        .catch(() => setAllowed(true));
    });
  }, []);

  if (allowed === null) {
    return (
      <div style={{ height: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'var(--hub-bg)' }}>
        <p style={{ color: 'var(--hub-text-muted)', fontSize: '14px' }}>Loading...</p>
      </div>
    );
  }

  if (!allowed) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh', background: 'var(--hub-bg)', color: 'var(--hub-text)', gap: '12px', fontSize: '14px' }}>
        <p style={{ fontWeight: 600 }}>You don't have access to this app</p>
        <p style={{ color: 'var(--hub-text-muted)', fontSize: '13px' }}>Contact your admin to get access.</p>
        <a href={APP_URLS.hub} style={{ color: '#4583C4', fontSize: '13px' }}>Go to Hub</a>
      </div>
    );
  }

  return <>{children}</>;
}

const App: React.FC = () => {
  return (
    <AuthGuard>
      <AppRoutes />
    </AuthGuard>
  );
};

export default App;
