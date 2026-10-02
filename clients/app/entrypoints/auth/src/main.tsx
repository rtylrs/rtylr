import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { FullScreenLoader } from '@voxire/components';
import { bootObservability } from '@voxire/observability';
import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/i18next';
import { Toaster } from 'sonner';
import { queryClient } from './shared/lib/queryClient';
import { router } from './router';

import 'react-lazy-load-image-component/src/effects/blur.css';
import './globals.css';

bootObservability({ app: 'auth' });

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <Suspense fallback={<FullScreenLoader />}>
      <I18nextProvider i18n={i18next}>
        <QueryClientProvider client={queryClient}>
          <RouterProvider router={router} />
          <Toaster expand={false} richColors />
        </QueryClientProvider>
      </I18nextProvider>
    </Suspense>
  );
} else {
  document.body.innerHTML = '<div>The App Failed To Load</div>';
}
