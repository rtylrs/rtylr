import React, { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { Toaster } from 'sonner';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { bootObservability } from '@voxire/observability';
import { loadRuntimeConfig } from '@voxire/config';
import { router } from './router';
import { initTheme } from './shared/stores/themeStore';
import './shared/services/i18next';
import './globals.css';

// Apply persisted theme before first render to prevent flash
initTheme();

bootObservability({ app: 'flow' });

const BootLoader = (): React.JSX.Element => (
  <div
    style={{
      display: 'flex',
      height: '100vh',
      alignItems: 'center',
      justifyContent: 'center',
      background: '#0a0a0a',
      color: 'rgba(240,240,238,0.5)',
      fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
      fontSize: '14px',
    }}
  >
    Loading Flow...
  </div>
);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60,
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  // Await runtime config (self-host) before rendering; loader meanwhile, static
  // hosted defaults on failure. Hosted is unchanged.
  root.render(<BootLoader />);
  void loadRuntimeConfig().finally(() => {
    root.render(
      <QueryClientProvider client={queryClient}>
        <Suspense fallback={<BootLoader />}>
          <RouterProvider router={router} />
          <Toaster expand={false} richColors />
        </Suspense>
      </QueryClientProvider>
    );
  });
}
