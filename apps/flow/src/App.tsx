import React from 'react';
import { useRoutes } from 'react-router-dom';
import { AppGuard } from '@voxire/auth';
import appRoutes from './routes/AppRoutes';
import api from './shared/lib/api';

const AppRouter: React.FC = (): React.JSX.Element => {
  const element = useRoutes(appRoutes);
  return <>{element}</>;
};

const App: React.FC = (): React.JSX.Element => {
  return (
    <AppGuard
      appId="flow"
      fetchUser={(): Promise<Record<string, unknown>> =>
        api.get('/user/me').then((r: { data: Record<string, unknown> }) => r.data)
      }
    >
      <AppRouter />
    </AppGuard>
  );
};

export default App;
