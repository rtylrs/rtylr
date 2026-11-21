import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { CookiesProvider } from 'react-cookie';
import { HelmetProvider } from 'react-helmet-async';
import { Toaster } from 'sonner';
import { injectTheme } from '@voxire/config';
import { bootObservability } from '@voxire/observability';
import '@voxire/ui/toast.css';
import i18next from '../../shared/services/i18next';
import { queryClient } from '../../shared/lib/react-query';
import { FullScreenLoader } from '../../shared/components/FullScreenLoader';
import { initTheme } from '../../shared/stores/themeStore';
import { renniliRouter } from './router';
import '../../globals.css';

// The staff pages reuse the Portal's components and theme tokens.
initTheme();
injectTheme();

bootObservability({ app: 'rennili' });

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <HelmetProvider>
      <QueryClientProvider client={queryClient}>
        <CookiesProvider>
          <Suspense fallback={<FullScreenLoader />}>
            <I18nextProvider i18n={i18next}>
              <RouterProvider router={renniliRouter} />
              <Toaster className="rtylr-toaster" expand={false} offset={20} mobileOffset={16} />
            </I18nextProvider>
          </Suspense>
        </CookiesProvider>
      </QueryClientProvider>
    </HelmetProvider>
  );
}
