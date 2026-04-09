import React, { useEffect, Suspense } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './screens/Extras/NoMatch';
import { useUserStore } from './stores/userStore';
import { useEmployeeSessionStore } from './stores/employeeSessionStore';
import PinLogin from './screens/Login';

const App = (): JSX.Element => {
  const location = useLocation();
  const { token, getUserInfo } = useUserStore();
  const { isActive } = useEmployeeSessionStore();

  // Restore business session on mount if token exists
  useEffect(() => {
    if (token) {
      getUserInfo().catch(() => {});
    }
  }, []);

  // Always require employee PIN — email/password login belongs in hub/auth
  if (!isActive) {
    return <PinLogin />;
  }

  return (
    <Suspense fallback={null}>
      <Routes key={location.pathname} location={location}>
        {AppRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
        <Route path="*" element={<NoMatch />} />
      </Routes>
    </Suspense>
  );
};

export default App;
