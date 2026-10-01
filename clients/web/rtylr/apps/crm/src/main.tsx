import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { Toaster } from 'sonner';
import { CookiesProvider } from 'react-cookie';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { bootObservability } from '@voxire/observability';
import { router } from './router';
import { readLocalStorage } from './shared/lib/storage';

import './shared/services/i18next';
import './globals.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60,
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

// Apply persisted theme before first render to prevent flash
const saved = readLocalStorage('crm-theme') || 'dark';
document.documentElement.setAttribute(
  'data-theme',
  saved === 'system'
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
    : saved
);

bootObservability({ app: 'crm' });

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <RouterProvider router={router} />
        <Toaster expand={false} richColors />
      </CookiesProvider>
    </QueryClientProvider>
  );
}
