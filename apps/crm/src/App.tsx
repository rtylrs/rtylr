import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import { ShortcutsProvider } from './shared/components/ShortcutsProvider';
import api from './shared/lib/api';

function NotFound(): React.JSX.Element {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: '#fff',
        color: '#111',
        gap: '8px',
      }}
    >
      <span style={{ fontSize: '48px', fontWeight: 700, color: '#4583C4' }}>404</span>
      <span style={{ fontSize: '14px', color: '#999' }}>Page not found</span>
      <a href="/" style={{ color: '#4583C4', fontSize: '13px', marginTop: 8 }}>
        Go home
      </a>
    </div>
  );
}

const AppRouter: React.FC = (): React.JSX.Element => (
  <ShortcutsProvider>
    <Routes>
      {AppRoutes.map((route, i) => (
        <Route key={i} path={route.path} element={route.element}>
          {route.children?.map((child, j) => (
            <Route key={j} index={child.index} path={child.path} element={child.element} />
          ))}
        </Route>
      ))}
      <Route path="*" element={<NotFound />} />
    </Routes>
  </ShortcutsProvider>
);

const App: React.FC = (): React.JSX.Element => (
  <ErrorBoundary>
    <AppGuard
      appId="crm"
      fetchUser={(): Promise<Record<string, unknown>> =>
        api.get('/user/me').then((r: { data: Record<string, unknown> }) => r.data)
      }
    >
      <AppFlagGuard flag={FLAG_KEYS.APP_CRM}>
        <AppRouter />
      </AppFlagGuard>
    </AppGuard>
  </ErrorBoundary>
);

export default App;
