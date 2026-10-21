import { Suspense } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { router } from './router';
import { FullScreenLoader } from './shared/components/FullScreenLoader';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/i18next';
import { Toaster } from 'sonner';
import '@voxire/ui/toast.css';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from './shared/lib/react-query';
import { HelmetProvider } from 'react-helmet-async';
import { bootObservability } from '@voxire/observability';
import HelmetContainer from './shared/components/HelmetContainer';
import { CookieBanner } from '@voxire/consent';

import './globals.css';
import 'react-lazy-load-image-component/src/effects/blur.css';
import { initTheme } from './shared/stores/themeStore';

// Apply persisted theme before first render to prevent flash
initTheme();

injectTheme();

bootObservability({ app: 'portal' });

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <HelmetProvider>
      <QueryClientProvider client={queryClient}>
        <CookiesProvider>
          <Suspense fallback={<FullScreenLoader />}>
            <I18nextProvider i18n={i18next}>
              <HelmetContainer title="Portal - RTYLR" favicon="/favicon.ico" />
              <CookieBanner />
              <RouterProvider router={router} />
              <Toaster className="rtylr-toaster" expand={false} offset={20} mobileOffset={16} />
            </I18nextProvider>
          </Suspense>
        </CookiesProvider>
      </QueryClientProvider>
    </HelmetProvider>
  );
} else {
  const msg = document.createElement('div');
  msg.textContent = 'The app failed to load. Please refresh the page.';
  msg.style.cssText = 'padding:2rem;font-family:sans-serif';
  document.body.appendChild(msg);
}
