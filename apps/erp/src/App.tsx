import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import React from 'react';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import AppRoutes from './routes/AppRoutes';
import api from './shared/lib/api';

const App: React.FC = (): React.JSX.Element => {
  return (
    <ErrorBoundary>
      <AppGuard
        appId="erp"
        fetchUser={(): Promise<Record<string, unknown>> =>
          api.get('/user/me').then((r: { data: Record<string, unknown> }) => r.data)
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
