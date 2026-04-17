import React from 'react';
import { AppGuard } from '@voxire/auth';
import AppRoutes from './routes/AppRoutes';
import api from './shared/lib/api';

const App: React.FC = (): React.JSX.Element => {
  return (
    <AppGuard
      appId="erp"
      fetchUser={(): Promise<Record<string, unknown>> =>
        api.get('/user/me').then((r: { data: Record<string, unknown> }) => r.data)
      }
    >
      <AppRoutes />
    </AppGuard>
  );
};

export default App;
