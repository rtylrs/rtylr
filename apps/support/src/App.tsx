import React, { useEffect, useState } from 'react';
import { APP_URLS } from '@voxire/config';
import AppRoutes from './routes/AppRoutes';
import api from './shared/lib/api';

const APP_ID = 'support';

function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(';').shift() ?? null;
  return null;
}

const ProtectedApp: React.FC = () => {
  const [allowed, setAllowed] = useState<boolean | null>(null);

  useEffect(() => {
    const token = getCookie('token');
    if (!token) {
      window.location.href = `${APP_URLS.auth}?redirect=${APP_ID}`;
      return;
    }

    api
      .get('/user/me')
      .then((res) => {
        const apps: string[] = res.data?.apps ?? [];
        const role: string = res.data?.role ?? '';
        if (role === 'owner' || role === 'admin' || apps.length === 0 || apps.includes(APP_ID)) {
          setAllowed(true);
        } else {
          setAllowed(false);
        }
      })
      .catch(() => {
        setAllowed(true);
      });
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

  if (allowed === false) {
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

  return <AppRoutes />;
};

const App: React.FC = () => {
  return <ProtectedApp />;
};

export default App;
