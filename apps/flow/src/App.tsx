import React, { useEffect } from 'react';
import { useRoutes, useNavigate } from 'react-router-dom';
import { getCookie } from './lib/utils';
import { useUserStore } from './stores/userStore';
import appRoutes from './routes/AppRoutes';

const AUTH_COOKIE_NAME = 'rtylr_token';
const AUTH_HUB_URL = 'http://localhost:5176';

const AuthGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, login } = useUserStore();

  useEffect(() => {
    const token = getCookie(AUTH_COOKIE_NAME);
    if (!token && !isAuthenticated) {
      // Redirect to hub auth with redirect_to=flow
      const redirectUrl = encodeURIComponent(window.location.href);
      window.location.href = `${AUTH_HUB_URL}?redirect=${redirectUrl}`;
      return;
    }

    if (token && !isAuthenticated) {
      // Hydrate a minimal user from token/cookie
      login(
        {
          id: 'local',
          name: 'Mohamed H',
          email: 'me@rtylr.com',
        },
        token
      );
    }
  }, [isAuthenticated, login]);

  if (!isAuthenticated && !getCookie(AUTH_COOKIE_NAME)) {
    return (
      <div
        style={{
          display: 'flex',
          height: '100vh',
          alignItems: 'center',
          justifyContent: 'center',
          background: '#0a0a0a',
          color: 'rgba(240,240,238,0.5)',
          fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
          fontSize: '14px',
        }}
      >
        Redirecting to Hub…
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
