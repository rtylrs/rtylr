import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Toaster } from 'sonner';
import { bootObservability } from '@voxire/observability';
import { loadRuntimeConfig } from '@voxire/config';
import { FullScreenLoader } from '@voxire/components';
import { router } from './router';
import './shared/services/i18next';
import './globals.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 1000 * 60 * 5,
    },
  },
});

bootObservability({ app: 'erp' });

const root = ReactDOM.createRoot(document.getElementById('root')!);
// Fetch runtime config (self-host URLs / feature posture) before rendering the
// router/guards; show the loader meanwhile and fall back to static hosted
// defaults if it fails. Hosted is unchanged.
root.render(<FullScreenLoader />);
void loadRuntimeConfig().finally(() => {
  root.render(
    <React.StrictMode>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
        <Toaster expand={false} richColors />
      </QueryClientProvider>
    </React.StrictMode>
  );
});
