import React, { useEffect, Suspense } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './screens/Extras/NoMatch';
import { useUserStore } from './stores/userStore';
import { useEmployeeSessionStore } from './stores/employeeSessionStore';
import PinLogin from './screens/Login';
import CommandPalette from './components/CommandPalette';
import HelpOverlay from './components/HelpOverlay';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useAutoLogout } from './hooks/useAutoLogout';

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
