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

/** Auth guard — checks for session cookie, redirects to auth app if missing */
function AuthGuard({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const [checking, setChecking] = useState(true);
  const [authed, setAuthed] = useState(false);

  useEffect(() => {
    // Check common session cookies used across rtylr apps
    const token =
      getCookie('rtylr_token') ||
      getCookie('rtylr_session') ||
      getCookie('auth_token') ||
      getCookie('session');

    if (token) {
      setAuthed(true);
    } else {
      // Redirect to auth app with redirect param so it can return here after login
      const redirect = encodeURIComponent('hr');
      window.location.href = `${APP_URLS.auth}?redirect=${redirect}`;
      return;
    }
    setChecking(false);
  }, []);

  if (checking && !authed) {
    return (
      <div
        style={{
          height: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'var(--hr-bg)',
        }}
      >
        <div
          style={{
            width: 32,
            height: 32,
            borderRadius: '50%',
            border: '2px solid rgba(139,92,246,0.2)',
            borderTopColor: '#8b5cf6',
            animation: 'spin 0.7s linear infinite',
          }}
        />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  if (!authed) return null;

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
