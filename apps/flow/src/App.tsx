import React from 'react';
import { useRoutes } from 'react-router-dom';
import { unwrap } from '@voxire/api';
import { AppGuard } from '@voxire/auth';
import appRoutes from './routes/AppRoutes';
import { safe } from './shared/lib/api';

const AppRouter: React.FC = (): React.JSX.Element => {
  const element = useRoutes(appRoutes);
  return <>{element}</>;
};

const App: React.FC = (): React.JSX.Element => {
  return (
    <AppGuard
      appId="flow"
      fetchUser={(): Promise<Record<string, unknown>> =>
        unwrap(safe.get<Record<string, unknown>>('/user/me'))
      }
    >
      <AppRouter />
    </AppGuard>
  );
};

export default App;
