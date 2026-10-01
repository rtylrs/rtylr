import React from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { Toaster } from 'sonner';
import { bootObservability } from '@voxire/observability';
import i18next from './shared/services/i18next';
import { router } from './router';
import './globals.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 1000 * 60 * 5,
    },
  },
});

bootObservability({ app: 'insights' });

const root = ReactDOM.createRoot(document.getElementById('root')!);
root.render(
  <React.StrictMode>
    <I18nextProvider i18n={i18next}>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
        <Toaster expand={false} richColors />
      </QueryClientProvider>
    </I18nextProvider>
  </React.StrictMode>
);
