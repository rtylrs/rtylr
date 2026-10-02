import { usePOSPortal } from './shared/components/PortalContext';
import React, { useEffect, Suspense, useRef, useState } from 'react';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { loadPOSSettings } from './shared/services/settings/loadPOSSettings';
import { ErrorBoundary, Spinner } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './features/extras/NoMatch';
import {
  useUserStore,
  useEmployeeSessionStore,
  useCartStore,
  useTerminalContextStore,
  useSettingsStore,
} from './shared/stores';
import { resolveAuthUrl } from './shared/services/server/serverConfig';
import PinLogin from './features/login';
import LocationSelect from './features/login/LocationSelect';
import CommandPalette from './shared/components/CommandPalette';
import HelpOverlay from './shared/components/HelpOverlay';
import {
  useKeyboardShortcuts,
  useAutoLogout,
  useOnlineStatus,
  useStockAlerts,
} from './shared/hooks';
import { restoreUserSessionFromCookie } from './shared/services/auth/sessionRestore';
import { useOfflineQueueAutoSync } from './features/offline-queue/useOfflineQueueAutoSync';
import { reportReactError } from './shared/services/observability/appErrors';
import { useTerminalHeartbeat } from './shared/services/terminal/useTerminalHeartbeat';
import { buildCartContext } from './shared/services/terminal/cartContext';
import { getCookie } from './shared/services/api/cookies';
import { runDesktopAutoUpdater } from './shared/services/desktop/autoUpdater';
import {
  businessAddress,
  businessContactPhone,
  businessDisplayName,
} from './shared/stores/userStore';

// Single shared cookie parser (the App.tsx copy broke on '='-padded JWTs
// and skipped decoding).
const getTokenFromCookie = (): string | null => getCookie('token');

const AppInner = (): React.JSX.Element => {
  const embedded = usePOSPortal();
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
      {!embedded && <CommandPalette />}
      <HelpOverlay />
    </Suspense>
  );
};

const AppShell = (): React.JSX.Element => {
  const embedded = usePOSPortal();
  const [settingsReady, setSettingsReady] = useState(false);
  const [settingsError, setSettingsError] = useState<Error | null>(null);
  const { token, setToken, getUserInfo } = useUserStore();
  const userBusinessProfile = useUserStore((s) => s.businessProfile);
  const { isActive, employee, endSession } = useEmployeeSessionStore();
  const setContext = useCartStore((s) => s.setContext);
  const setBusiness = useSettingsStore((s) => s.setBusiness);
  const { locationId, terminalId, setTerminalContext } = useTerminalContextStore();
  const hasRestoredSession = useRef(false);

  useEffect(() => {
    runDesktopAutoUpdater();
  }, []);

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
        window.location.href = `${resolveAuthUrl()}?redirect=pos`;
      },
    });
  }, [getUserInfo, setToken, token]);

  useEffect(() => {
    const name = businessDisplayName(userBusinessProfile);
    if (!name) return;

    setBusiness({
      name,
      address: businessAddress(userBusinessProfile) ?? '',
      phone: businessContactPhone(userBusinessProfile) ?? '',
    });
  }, [setBusiness, userBusinessProfile]);

  useEffect(() => {
    if (!token && !getTokenFromCookie()) return;

    let active = true;
    setSettingsReady(false);
    setSettingsError(null);
    void loadPOSSettings(() => active)
      .then(() => {
        if (active) setSettingsReady(true);
      })
      .catch((error: unknown) => {
        if (!active) return;
        setSettingsError(
          error instanceof Error ? error : new Error('Could not load terminal configuration')
        );
      });
    return () => {
      active = false;
    };
  }, [token]);

  useEffect(() => {
    if (!locationId || !terminalId || !isActive || !employee) return;

    const context = buildCartContext(locationId, terminalId, employee.id);
    if (!context) {
      endSession();
      return;
    }

    setContext(context.branchId, context.terminalId, context.employeeId);
  }, [employee, endSession, isActive, locationId, setContext, terminalId]);

  if (embedded && settingsError) throw settingsError;
  if (embedded && !settingsReady)
    return (
      <div className="flex min-h-48 items-center justify-center">
        <Spinner />
      </div>
    );

  // No location/terminal selected → show LocationSelect.
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
