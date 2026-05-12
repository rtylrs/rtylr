import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import React from 'react';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import { unwrap } from '@voxire/api';
import { safe } from './shared/lib/api';

const App: React.FC = (): React.JSX.Element => {
  return (
    <ErrorBoundary>
      <AppGuard
        appId="erp"
        fetchUser={(): Promise<Record<string, unknown>> =>
          unwrap(safe.get<Record<string, unknown>>('/user/me'))
        }
      >
        <AppFlagGuard flag={FLAG_KEYS.APP_ERP}>
          <AppRoutes />
        </AppFlagGuard>
      </AppGuard>
    </ErrorBoundary>
  );
};

export default App;
