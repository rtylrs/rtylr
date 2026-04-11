import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { Toaster } from 'sonner';
import { CookiesProvider } from 'react-cookie';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

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
const saved = localStorage.getItem('crm-theme') || 'dark';
document.documentElement.setAttribute(
  'data-theme',
  saved === 'system'
    ? window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light'
    : saved
);

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
      Loading...
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
            <Toaster expand={false} richColors />
          </BrowserRouter>
        </Suspense>
      </CookiesProvider>
    </QueryClientProvider>
  );
}
