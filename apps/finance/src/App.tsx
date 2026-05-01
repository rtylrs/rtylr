import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import React from 'react';
import { Routes, Route, useLocation } from 'react-router-dom';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import api from './shared/lib/api';

const AppRouter: React.FC = (): React.JSX.Element => {
  const location = useLocation();
  return (
    <Routes key={location.pathname} location={location}>
      {AppRoutes.map((route, i) => (
        <Route key={i} path={route.path} element={route.element}>
          {route.children?.map((child, j) => (
            <Route key={j} index={child.index} path={child.path} element={child.element} />
          ))}
        </Route>
      ))}
    </Routes>
  );
};

const App: React.FC = (): React.JSX.Element => (
  <ErrorBoundary>
    <AppGuard
      appId="finance"
      fetchUser={(): Promise<Record<string, unknown>> =>
        api.get('/user/me').then((r: { data: Record<string, unknown> }) => r.data)
      }
    >
      <AppFlagGuard flag={FLAG_KEYS.APP_FINANCE}>
        <AppRouter />
      </AppFlagGuard>
    </AppGuard>
  </ErrorBoundary>
);

export default App;
