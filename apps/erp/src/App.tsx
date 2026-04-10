import React, { useEffect, useState } from 'react';
import { APP_URLS } from '@voxire/config';
import AppRoutes from './routes/AppRoutes';
import api from './lib/api';

const APP_ID = 'erp';

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

    // Check if user has access to this app
    api.get('/user/me')
      .then((res) => {
        const apps: string[] = res.data?.apps ?? [];
        const role: string = res.data?.role ?? '';
        // Owners/admins or users with empty apps array (owner account) have full access
        if (role === 'owner' || role === 'admin' || apps.length === 0 || apps.includes(APP_ID)) {
          setAllowed(true);
        } else {
          setAllowed(false);
        }
      })
      .catch(() => {
        // Can't verify permissions, allow access (token might still be valid)
        setAllowed(true);
      });
  }, []);

  if (allowed === null) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh', background: 'var(--hub-bg)', color: 'var(--hub-text-muted)', fontSize: '14px' }}>
        Loading...
      </div>
    );
  }

  if (allowed === false) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh', background: 'var(--hub-bg)', color: 'var(--hub-text)', gap: '12px', fontSize: '14px' }}>
        <p style={{ fontWeight: 600 }}>You don't have access to this app</p>
        <p style={{ color: 'var(--hub-text-muted)', fontSize: '13px' }}>Contact your admin to get access.</p>
        <a href={APP_URLS.hub} style={{ color: '#4583C4', fontSize: '13px' }}>Go to Hub</a>
      </div>
    );
  }

  return <AppRoutes />;
};

const App: React.FC = () => {
  return <ProtectedApp />;
};

export default App;
