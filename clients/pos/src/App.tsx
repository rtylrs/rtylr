import React, { useEffect, Suspense, useRef } from 'react';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import { createOneKitClient } from '@voxire/api';
import { SettingsServiceClient } from '@voxire/voxie/pos/settings/v1/client';
import { ErrorBoundary } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import NoMatch from './features/extras/NoMatch';
import {
  useUserStore,
  useEmployeeSessionStore,
  useCartStore,
  useTerminalContextStore,
  useSettingsStore,
  useThemeStore,
  useServerStore,
} from './shared/stores';
import ServerSetup from './features/server-setup';
import { resolveAuthUrl } from './shared/services/server/serverConfig';
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
import { getCookie } from './shared/services/api/cookies';
import { runDesktopAutoUpdater } from './shared/services/desktop/autoUpdater';
import {
  businessAddress,
  businessContactPhone,
  businessDisplayName,
} from './shared/stores/userStore';
import type { POSBusinessProfile } from './shared/stores/settingsStore';

// Single shared cookie parser (the App.tsx copy broke on '='-padded JWTs
// and skipped decoding).
const getTokenFromCookie = (): string | null => getCookie('token');

const settingsClient = createOneKitClient(SettingsServiceClient, 'pos');

function isPOSBusinessProfile(value: string): value is POSBusinessProfile {
  return value === 'retail' || value === 'restaurant' || value === 'quick_service';
}

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
  const userBusinessProfile = useUserStore((s) => s.businessProfile);
  const { isActive, employee, endSession } = useEmployeeSessionStore();
  const setContext = useCartStore((s) => s.setContext);
  const setBusiness = useSettingsStore((s) => s.setBusiness);
  const setPOSBusinessProfile = useSettingsStore((s) => s.setBusinessProfile);
  const setPOSFeatures = useSettingsStore((s) => s.setFeatures);
  const setReceiptFooter = useSettingsStore((s) => s.setReceiptFooter);
  const setTheme = useThemeStore((s) => s.setTheme);
  const { locationId, terminalId, setTerminalContext } = useTerminalContextStore();
  const serverConfigured = useServerStore((s) => s.configured);
  const hasRestoredSession = useRef(false);
  const hasLoadedPOSSettings = useRef(false);

  useEffect(() => {
    runDesktopAutoUpdater();
  }, []);

  // Check auth — only after the backend has been chosen, so the very first
  // redirect targets the correct (hosted or self-host) auth host.
  useEffect(() => {
    if (!serverConfigured) return;
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
  }, [getUserInfo, serverConfigured, setToken, token]);

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
    if (hasLoadedPOSSettings.current) return;
    if (!token && !getTokenFromCookie()) return;

    hasLoadedPOSSettings.current = true;
    void settingsClient
      .getPOSSettings({})
      .then((data) => {
        const settings = data.settings;
        if (!settings) return;
        const businessProfile = settings.businessProfile ?? '';
        if (isPOSBusinessProfile(businessProfile)) {
          setPOSBusinessProfile(businessProfile);
        }
        if (settings.features) {
          setPOSFeatures(settings.features);
        }
        setReceiptFooter(settings.receiptFooter ?? '');
        setTheme(settings.theme ?? '');
      })
      .catch(() => {
        hasLoadedPOSSettings.current = false;
      });
  }, [setPOSBusinessProfile, setPOSFeatures, setReceiptFooter, setTheme, token]);

  useEffect(() => {
    if (!locationId || !terminalId || !isActive || !employee) return;

    const context = buildCartContext(locationId, terminalId, employee.id);
    if (!context) {
      endSession();
      return;
    }

    setContext(context.branchId, context.terminalId, context.employeeId);
  }, [employee, endSession, isActive, locationId, setContext, terminalId]);

  // Step 0: No backend chosen yet → first-run server chooser.
  if (!serverConfigured) {
    return <ServerSetup variant="first-run" />;
  }

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
