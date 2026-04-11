import React, { useEffect, Suspense } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './features/extras/NoMatch';
import { useUserStore } from './shared/stores/userStore';
import { useEmployeeSessionStore } from './shared/stores/employeeSessionStore';
import PinLogin from './features/login';
import CommandPalette from './shared/components/CommandPalette';
import HelpOverlay from './shared/components/HelpOverlay';
import { useKeyboardShortcuts } from './shared/hooks/useKeyboardShortcuts';
import { useAutoLogout } from './shared/hooks/useAutoLogout';

const AppInner = (): JSX.Element => {
  useKeyboardShortcuts();
  useAutoLogout();
  const location = useLocation();
  return (
    <Suspense fallback={null}>
      <Routes key={location.pathname} location={location}>
        {AppRoutes.map((route, index) => (
          <Route key={index} path={route.path} element={route.element} />
        ))}
        <Route path="*" element={<NoMatch />} />
      </Routes>
      <CommandPalette />
      <HelpOverlay />
    </Suspense>
  );
};

const App = (): JSX.Element => {
  const { token, getUserInfo } = useUserStore();
  const { isActive } = useEmployeeSessionStore();

  useEffect(() => {
    if (token) {
      getUserInfo().catch(() => {});
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!isActive) {
    return <PinLogin />;
  }

  return <AppInner />;
};

export default App;
