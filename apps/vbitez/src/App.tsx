import React, { useEffect, Suspense, useState } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import { APP_URLS, AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { ErrorBoundary } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './features/extras/NoMatch';
import { useUserStore } from './shared/stores';
import { useEmployeeSessionStore } from './shared/stores/employeeSessionStore';
import { useCartStore } from './shared/stores/cartStore';
import PinLogin from './features/login';
import LocationSelect from './features/login/LocationSelect';
import CommandPalette from './shared/components/CommandPalette';
import HelpOverlay from './shared/components/HelpOverlay';
import { useKeyboardShortcuts } from './shared/hooks';
import { useAutoLogout } from './shared/hooks';

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const AppInner = (): React.JSX.Element => {
  useKeyboardShortcuts();
  useAutoLogout();
  const location = useLocation();
  return (
    <Suspense fallback={null}>
      <AppFlagGuard flag={FLAG_KEYS.APP_POS}>
        <Routes key={location.pathname} location={location}>
          {AppRoutes.map((route, index) => (
            <Route key={index} path={route.path} element={route.element} />
          ))}
          <Route path="*" element={<NoMatch />} />
        </Routes>
      </AppFlagGuard>
      <CommandPalette />
      <HelpOverlay />
    </Suspense>
  );
};

const AppShell = (): React.JSX.Element => {
  const { token, getUserInfo } = useUserStore();
  const { isActive } = useEmployeeSessionStore();
  const setContext = useCartStore((s) => s.setContext);
  const [locationId, setLocationId] = useState<number | null>(null);
  const [terminalId, setTerminalId] = useState<number | null>(null);

  // Check auth
  useEffect(() => {
    const cookieToken = getTokenFromCookie();
    if (!cookieToken) {
      window.location.href = `${APP_URLS.auth}?redirect=pos`;
      return;
    }
    if (token) {
      getUserInfo().catch(() => {});
    }
  }, []);

  // Step 1: No location/terminal selected → show LocationSelect
  if (!locationId || !terminalId) {
    return (
      <LocationSelect
        onComplete={(locId, termId) => {
          setLocationId(locId);
          setTerminalId(termId);
        }}
      />
    );
  }

  // Step 2: Location + terminal set but no employee session → show PinLogin
  if (!isActive) {
    return <PinLogin />;
  }

  // Step 3: Set cart context with location + terminal + employee
  const employee = useEmployeeSessionStore.getState().employee;
  if (employee) {
    setContext(locationId as any, terminalId as any, employee.id as any);
  }

  // Step 4: Everything ready → main POS app
  return <AppInner />;
};

const App = (): React.JSX.Element => (
  <ErrorBoundary>
    <AppShell />
  </ErrorBoundary>
);

export default App;
