import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { Toaster } from 'sonner';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { bootObservability } from '@voxire/observability';

import './shared/services/i18next';
import './globals.css';

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
  ReactDOM.createRoot(rootElement).render(
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <Suspense
          fallback={
            <div
              style={{
                display: 'flex',
                height: '100vh',
                alignItems: 'center',
                justifyContent: 'center',
                background: '#0a0a0a',
              }}
            >
              <div
                style={{
                  width: 32,
                  height: 32,
                  border: '2px solid rgba(16,185,129,0.3)',
                  borderTopColor: '#10b981',
                  borderRadius: '50%',
                  animation: 'spin 0.7s linear infinite',
                }}
              />
              <style>{`@keyframes spin { to { transform: rotate(360deg) } }`}</style>
            </div>
          }
        >
          <BrowserRouter>
            <App />
            <Toaster expand={false} richColors />
          </BrowserRouter>
        </Suspense>
      </CookiesProvider>
    </QueryClientProvider>
  );
}
