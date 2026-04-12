import React, { useEffect, useState } from 'react';
import { Routes, Route, useLocation } from 'react-router-dom';
import { useCookies } from 'react-cookie';
import { APP_URLS } from '@voxire/config';
import AppRoutes from './routes/AppRoutes';
import api from './shared/lib/api';

const APP_ID = 'finance';
const AUTH_COOKIE = 'token';

function AuthGuard({ children }: { children: React.ReactNode }) {
  const [cookies] = useCookies([AUTH_COOKIE]);
  const [allowed, setAllowed] = useState<boolean | null>(null);

  useEffect(() => {
    const token = cookies[AUTH_COOKIE];
    if (!token) {
      window.location.href = `${APP_URLS.auth}?redirect=${APP_ID}`;
      return;
    }

    api
      .get('/user/me')
      .then((res: { data?: { apps?: string[]; role?: string } }) => {
        const apps: string[] = res.data?.apps ?? [];
        const role: string = res.data?.role ?? '';
        if (role === 'owner' || role === 'admin' || apps.length === 0 || apps.includes(APP_ID)) {
          setAllowed(true);
        } else {
          setAllowed(false);
        }
      })
      .catch(() => setAllowed(true));
  }, [cookies]);

  if (allowed === null) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          background: '#fff',
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
        <style>{'@keyframes spin { to { transform: rotate(360deg) } }'}</style>
      </div>
    );
  }

  if (!allowed) {
    return (
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          background: '#fff',
          color: '#111',
          gap: '12px',
          fontSize: '14px',
        }}
      >
        <p style={{ fontWeight: 600 }}>You don't have access to this app</p>
        <p style={{ color: '#999', fontSize: '13px' }}>Contact your admin to get access.</p>
        <a href={APP_URLS.hub} style={{ color: '#4583C4', fontSize: '13px' }}>
          Go to Hub
        </a>
      </div>
    );
  }

  return <>{children}</>;
}

const App: React.FC = () => {
  const location = useLocation();

  return (
    <AuthGuard>
      <Routes key={location.pathname} location={location}>
        {AppRoutes.map((route, i) => (
          <Route key={i} path={route.path} element={route.element}>
            {route.children?.map((child, j) => (
              <Route key={j} index={child.index} path={child.path} element={child.element} />
            ))}
          </Route>
        ))}
      </Routes>
    </AuthGuard>
  );
};

export default App;
