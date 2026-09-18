import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { Toaster } from 'sonner';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { bootObservability } from '@voxire/observability';
import { loadRuntimeConfig } from '@voxire/config';
import { FullScreenLoader } from '@voxire/components';
import { router } from './router';

import './shared/services/i18next';
import './globals.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

bootObservability({ app: 'recruit' });

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  // Await runtime config (self-host) before rendering; loader meanwhile, static
  // hosted defaults on failure. Hosted is unchanged.
  root.render(<FullScreenLoader />);
  void loadRuntimeConfig().finally(() => {
    root.render(
      <QueryClientProvider client={queryClient}>
        <CookiesProvider>
          <RouterProvider router={router} />
          <Toaster expand={false} richColors position="top-center" />
        </CookiesProvider>
      </QueryClientProvider>
    );
  });
}
