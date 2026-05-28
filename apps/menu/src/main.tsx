import { Suspense, useEffect } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { FullScreenLoader } from '@voxire/components';
import { bootObservability } from '@voxire/observability';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useConfigStore } from './stores/configStore';
import { router } from './router';
import './globals.css';

const queryClient = new QueryClient();

injectTheme();

bootObservability({ app: 'menu' });

function AppInit(): null {
  const fetchConfigData = useConfigStore((s) => s.fetchConfigData);
  useEffect(() => {
    fetchConfigData();
  }, []);
  return null;
}

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense fallback={<FullScreenLoader />}>
      <QueryClientProvider client={queryClient}>
        <I18nextProvider i18n={i18next}>
          <AppInit />
          <RouterProvider router={router} />
        </I18nextProvider>
      </QueryClientProvider>
    </Suspense>
  );
}
