import React, { useEffect, useState } from 'react';
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom';
import { useCookies } from 'react-cookie';
import { APP_URLS } from '@voxire/config';
import AppRoutes from './routes/AppRoutes';
import { ShortcutsProvider } from './components/ShortcutsProvider';

const AUTH_COOKIE = 'token';

function AuthGuard({ children }: { children: React.ReactNode }) {
  const [cookies] = useCookies([AUTH_COOKIE]);
  const navigate = useNavigate();
  const location = useLocation();
  const [checked, setChecked] = useState(false);

  useEffect(() => {
    const token = cookies[AUTH_COOKIE];
    if (!token) {
      const redirectUrl = `${APP_URLS.auth}?redirect=crm&next=${encodeURIComponent(location.pathname)}`;
      window.location.href = redirectUrl;
    } else {
      setChecked(true);
    }
  }, [cookies, location.pathname]);

  if (!checked) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          background: '#0a0a0a',
          color: 'rgba(240,240,238,0.5)',
          fontFamily: "'Plus Jakarta Sans', system-ui, sans-serif",
          fontSize: '14px',
        }}
      >
        Authenticating...
      </div>
    );
  }

  return <>{children}</>;
}

function NotFound() {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: '#0a0a0a',
        color: 'rgba(240,240,238,0.92)',
        fontFamily: "'Plus Jakarta Sans', system-ui, sans-serif",
        gap: '8px',
      }}
    >
      <span style={{ fontSize: '48px', fontWeight: 700, color: '#3b82f6' }}>404</span>
      <span style={{ fontSize: '16px', color: 'rgba(240,240,238,0.5)' }}>Page not found</span>
    </div>
  );
}

const App: React.FC = () => {
  return (
    <AuthGuard>
      <ShortcutsProvider>
        <Routes>
          {AppRoutes.map((route, i) => (
            <Route key={i} path={route.path} element={route.element}>
              {route.children?.map((child, j) => (
                <Route key={j} index={child.index} path={child.path} element={child.element} />
              ))}
            </Route>
          ))}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </ShortcutsProvider>
    </AuthGuard>
  );
};

export default App;
