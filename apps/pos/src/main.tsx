import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { Toaster } from 'sonner';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';

import './globals.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

function FullScreenLoader() {
  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: '#0a0a0a',
        color: 'rgba(240,240,238,0.5)',
        fontFamily: "'Plus Jakarta Sans', system-ui, sans-serif",
        fontSize: '14px',
      }}
    >
      Loading POS…
    </div>
  );
}

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <Suspense fallback={<FullScreenLoader />}>
          <BrowserRouter>
            <App />
            <Toaster expand={false} richColors position="top-right" />
          </BrowserRouter>
        </Suspense>
      </CookiesProvider>
    </QueryClientProvider>
  );
}
