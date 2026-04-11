import React, { useEffect, useState } from 'react';
import { useRoutes } from 'react-router-dom';
import { APP_URLS } from '@voxire/config';
import { getCookie } from './shared/lib/utils';
import { useUserStore } from './shared/stores/userStore';
import appRoutes from './routes/AppRoutes';
import api from './shared/lib/api';
import { mapMeResponseToFlowUser } from './shared/lib/mapMeUser';

const APP_ID = 'flow';

const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [allowed, setAllowed] = useState<boolean | null>(null);

  useEffect(() => {
    const token = getCookie('token');
    if (!token) {
      window.location.href = `${APP_URLS.auth}?redirect=${APP_ID}`;
      return;
    }

    api
      .get('/user/me')
      .then((res: { data?: Record<string, unknown> }) => {
        const payload = res.data;
        const mapped = mapMeResponseToFlowUser(payload);
        if (mapped) {
          // Always refresh profile from API (Hub-style) so persisted zustand state does not stay stale.
          useUserStore.getState().login(mapped, token);
        }

        const apps: string[] = (payload?.apps as string[]) ?? [];
        const role: string = (payload?.role as string) ?? '';

        if (role === 'owner' || role === 'admin' || apps.length === 0 || apps.includes(APP_ID)) {
          setAllowed(true);
        } else {
          setAllowed(false);
        }
      })
      .catch(() => setAllowed(true));
  }, []);

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
};

const AppRouter: React.FC = () => {
  const element = useRoutes(appRoutes);
  return element;
};

const App: React.FC = () => {
  return (
    <AuthGuard>
      <AppRouter />
    </AuthGuard>
  );
};

export default App;
