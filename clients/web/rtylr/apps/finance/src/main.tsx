import React, { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { router } from './router';
import { Toaster } from 'sonner';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { bootObservability } from '@voxire/observability';

import './shared/services/i18next';
import './globals.css';

const BootLoader = (): React.JSX.Element => (
  <div
    style={{
      display: 'flex',
      height: '100vh',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'var(--hub-bg)',
    }}
  >
    <div
      style={{
        width: 32,
        height: 32,
        border: '2px solid var(--finance-accent-border)',
        borderTopColor: 'var(--finance-accent)',
        borderRadius: '50%',
        animation: 'spin 0.7s linear infinite',
      }}
    />
    <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
  </div>
);

// Apply persisted theme before first render to prevent flash
const saved = localStorage.getItem('finance-theme') || 'dark';
document.documentElement.setAttribute(
  'data-theme',
  saved === 'system'
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
    : saved
);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

bootObservability({ app: 'finance' });

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <Suspense fallback={<BootLoader />}>
          <RouterProvider router={router} />
          <Toaster expand={false} richColors />
        </Suspense>
      </CookiesProvider>
    </QueryClientProvider>
  );
}
