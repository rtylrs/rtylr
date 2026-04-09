import React, { useEffect, useState } from 'react';
import { Routes, Route, useLocation, useNavigate } from 'react-router-dom';
import { useCookies } from 'react-cookie';
import { APP_URLS } from '@voxire/config';
import AppRoutes from './routes/AppRoutes';

const AUTH_COOKIE = 'rtylr_session';

function useAuthGuard() {
  const [cookies] = useCookies([AUTH_COOKIE]);
  const navigate = useNavigate();
  const [checked, setChecked] = useState(false);
  const [authed, setAuthed] = useState(false);

  useEffect(() => {
    const session = cookies[AUTH_COOKIE];
    if (!session) {
      // Redirect to auth with redirect param pointing back to finance
      const redirectTarget = encodeURIComponent(APP_URLS.finance);
      window.location.href = `${APP_URLS.auth}?redirect=${redirectTarget}`;
    } else {
      setAuthed(true);
    }
    setChecked(true);
  }, [cookies]);

  return { checked, authed };
}

const App: React.FC = () => {
  const location = useLocation();
  const { checked, authed } = useAuthGuard();

  if (!checked) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', background: '#0a0a0a' }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ width: 32, height: 32, border: '2px solid rgba(16,185,129,0.3)', borderTopColor: '#10b981', borderRadius: '50%', animation: 'spin 0.7s linear infinite', margin: '0 auto 12px' }} />
          <p style={{ color: 'rgba(240,240,238,0.5)', fontSize: 13, margin: 0 }}>Verifying session…</p>
          <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
        </div>
      </div>
    );
  }

  if (!authed) return null;

  return (
    <Routes key={location.pathname} location={location}>
      {AppRoutes.map((route, i) => (
        <Route key={i} path={route.path} element={route.element}>
          {route.children?.map((child, j) => (
            <Route key={j} index={child.index} path={child.path} element={child.element} />
          ))}
        </Route>
      ))}
    </Routes>
  );
};

export default App;
