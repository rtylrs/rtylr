import React, { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { Toaster } from 'sonner';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { bootObservability } from '@voxire/observability';

import App from './App';
import './globals.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

bootObservability({ app: 'recruit' });

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
                background: 'var(--recruit-bg)',
              }}
            >
              <div
                style={{
                  width: 28,
                  height: 28,
                  border: '2px solid rgba(17,17,17,0.15)',
                  borderTopColor: 'var(--recruit-accent)',
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
            <Toaster expand={false} richColors position="top-center" />
          </BrowserRouter>
        </Suspense>
      </CookiesProvider>
    </QueryClientProvider>
  );
}
