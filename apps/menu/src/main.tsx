import { Suspense } from 'react';
import { injectTheme, loadRuntimeConfig } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { FullScreenLoader } from '@voxire/components';
import { bootObservability } from '@voxire/observability';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { router } from './router';
import './globals.css';

const queryClient = new QueryClient();

injectTheme();

bootObservability({ app: 'menu' });

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  // Await runtime config (self-host) before rendering; loader meanwhile, static
  // hosted defaults on failure. Hosted is unchanged.
  root.render(<FullScreenLoader />);
  void loadRuntimeConfig().finally(() => {
    root.render(
      <Suspense fallback={<FullScreenLoader />}>
        <QueryClientProvider client={queryClient}>
          <I18nextProvider i18n={i18next}>
            <RouterProvider router={router} />
          </I18nextProvider>
        </QueryClientProvider>
      </Suspense>
    );
  });
}
