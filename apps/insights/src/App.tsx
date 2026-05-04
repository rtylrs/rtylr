import { AppFlagGuard, FLAG_KEYS } from '@voxire/config';
import React from 'react';
import { AppGuard } from '@voxire/auth';
import { ErrorBoundary } from '@voxire/ui';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Filler,
  Tooltip,
  Legend,
} from 'chart.js';
import AppRoutes from './routes/AppRoutes';
import api from './shared/lib/api';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  Filler,
  Tooltip,
  Legend
);

const App: React.FC = (): React.JSX.Element => {
  return (
    <ErrorBoundary>
      <AppGuard
        appId="insights"
        fetchUser={(): Promise<Record<string, unknown>> =>
          api.get('/user/me').then((r: { data: Record<string, unknown> }) => r.data)
        }
      >
        <AppFlagGuard flag={FLAG_KEYS.APP_INSIGHTS}>
          <AppRoutes />
        </AppFlagGuard>
      </AppGuard>
    </ErrorBoundary>
  );
};

export default App;
