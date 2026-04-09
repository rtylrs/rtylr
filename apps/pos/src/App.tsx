import React, { useEffect } from 'react';
import { Routes, Route, useNavigate } from 'react-router-dom';
import { useCookies } from 'react-cookie';
import AppRoutes from './routes/AppRoutes';

/**
 * Auth guard: if no `auth_token` cookie is present,
 * redirect to the Hub login page with ?redirect=pos so
 * the user is bounced back here after authenticating.
 */
function AuthGuard({ children }: { children: React.ReactNode }) {
  const [cookies] = useCookies(['auth_token']);
  const navigate = useNavigate();

  useEffect(() => {
    if (!cookies.auth_token) {
      // Hub login lives on port 5176 in local dev
      window.location.href = 'http://localhost:5176/auth/login?redirect=pos';
    }
  }, [cookies.auth_token]);

  if (!cookies.auth_token) return null;

  return <>{children}</>;
}

const App: React.FC = () => {
  return (
    <AuthGuard>
      <Routes>
        {AppRoutes.map((route, i) => (
          <Route key={i} path={route.path} element={route.element} />
        ))}
      </Routes>
    </AuthGuard>
  );
};

export default App;
