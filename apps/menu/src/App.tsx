import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import React, { Suspense, useEffect } from 'react';
import { Route, Routes, useLocation } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import { ErrorBoundary } from '@voxire/ui';
import PublicRoutes from './routes/PublicRoutes';
import { useConfigStore } from './stores/configStore';
import '@voxire/menu/styles';

const App = (): JSX.Element => {
  const location = useLocation();
  const fetchConfigData = useConfigStore((s) => s.fetchConfigData);

  useEffect(() => {
    fetchConfigData();
  }, []);

  return (
    <ErrorBoundary>
      <Suspense fallback={<FullScreenLoader />}>
        <AppFlagGuard flag={FLAG_KEYS.APP_MENU}>
          <Routes key={location.pathname} location={location}>
            {PublicRoutes.map((route, index) => (
              <Route key={index} path={route.path} element={route.element} />
            ))}
          </Routes>
        </AppFlagGuard>
      </Suspense>
    </ErrorBoundary>
  );
};

export default App;
