import React, { useEffect, Suspense, useRef } from 'react';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { APP_URLS, AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { ErrorBoundary } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './features/extras/NoMatch';
import {
  useUserStore,
  useEmployeeSessionStore,
  useCartStore,
  useTerminalContextStore,
  useSettingsStore,
} from './shared/stores';
import PinLogin from './features/login';
import LocationSelect from './features/login/LocationSelect';
import CommandPalette from './shared/components/CommandPalette';
import HelpOverlay from './shared/components/HelpOverlay';
import { useKeyboardShortcuts, useAutoLogout, useOnlineStatus, useStockAlerts } from './shared/hooks';
import { restoreUserSessionFromCookie } from './shared/services/auth/sessionRestore';
import { useOfflineQueueAutoSync } from './features/offline-queue/useOfflineQueueAutoSync';
import { reportReactError } from './shared/services/observability/appErrors';
import { useTerminalHeartbeat } from './shared/services/terminal/useTerminalHeartbeat';
import { buildCartContext } from './shared/services/terminal/cartContext';
import {
  businessAddress,
  businessContactPhone,
  businessDisplayName,
} from './shared/stores/userStore';

const getTokenFromCookie = (): string | null =>
  document.cookie
    .split(';')
    .map((c) => c.trim())
    .find((c) => c.startsWith('token='))
    ?.split('=')[1] ?? null;

const AppInner = (): React.JSX.Element => {
  useKeyboardShortcuts();
  useAutoLogout();
  useOnlineStatus();
  useOfflineQueueAutoSync();
  useTerminalHeartbeat(true);
  useStockAlerts();
  const location = useLocation();
  const features = useSettingsStore((s) => s.features);
  return (
    <Suspense fallback={null}>
      <AppFlagGuard flag={FLAG_KEYS.APP_POS}>
        <Routes key={location.pathname} location={location}>
          {AppRoutes.map((route, index) => (
            <Route
              key={index}
              path={route.path}
              element={
                route.feature && !features[route.feature] ? (
                  <Navigate to="/" replace />
                ) : (
                  route.element
                )
              }
            />
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
  const { token, setToken, getUserInfo } = useUserStore();
  const businessProfile = useUserStore((s) => s.businessProfile);
  const { isActive, employee, endSession } = useEmployeeSessionStore();
  const setContext = useCartStore((s) => s.setContext);
  const setBusiness = useSettingsStore((s) => s.setBusiness);
  const { locationId, terminalId, setTerminalContext } = useTerminalContextStore();
  const hasRestoredSession = useRef(false);

  // Check auth
  useEffect(() => {
    if (hasRestoredSession.current) return;
    hasRestoredSession.current = true;

    const cookieToken = getTokenFromCookie();
    restoreUserSessionFromCookie({
      cookieToken,
      storeToken: token,
      setToken,
      getUserInfo,
      redirectToAuth: () => {
        window.location.href = `${APP_URLS.auth}?redirect=pos`;
      },
    });
  }, [getUserInfo, setToken, token]);

  useEffect(() => {
    const name = businessDisplayName(businessProfile);
    if (!name) return;

    setBusiness({
      name,
      address: businessAddress(businessProfile) ?? '',
      phone: businessContactPhone(businessProfile) ?? '',
    });
  }, [businessProfile, setBusiness]);

  useEffect(() => {
    if (!locationId || !terminalId || !isActive || !employee) return;

    const context = buildCartContext(locationId, terminalId, employee.id);
    if (!context) {
      endSession();
      return;
    }

    setContext(context.branchId, context.terminalId, context.employeeId);
  }, [employee, endSession, isActive, locationId, setContext, terminalId]);

  // Step 1: No location/terminal selected → show LocationSelect
  if (!locationId || !terminalId) {
    return (
      <LocationSelect
        onComplete={(locId, termId) => {
          setTerminalContext(locId, termId);
        }}
      />
    );
  }

  // Step 2: Location + terminal set but no employee session → show PinLogin
  if (!isActive) {
    return <PinLogin />;
  }

  // Step 3: Everything ready → main POS app
  return <AppInner />;
};

const App = (): React.JSX.Element => (
  <ErrorBoundary onError={reportReactError}>
    <AppShell />
  </ErrorBoundary>
);

export default App;
