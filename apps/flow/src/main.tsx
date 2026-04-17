import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { Toaster } from 'sonner';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App';
import { initTheme } from './shared/stores/themeStore';
import './shared/services/i18next';
import './globals.css';

// Apply persisted theme before first render to prevent flash
initTheme();

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60,
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <QueryClientProvider client={queryClient}>
      <Suspense
        fallback={
          <div
            style={{
              display: 'flex',
              height: '100vh',
              alignItems: 'center',
              justifyContent: 'center',
              background: '#0a0a0a',
              color: 'rgba(240,240,238,0.5)',
              fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
              fontSize: '14px',
            }}
          >
            Loading Flow…
          </div>
        }
      >
        <BrowserRouter>
          <App />
          <Toaster expand={false} richColors />
        </BrowserRouter>
      </Suspense>
    </QueryClientProvider>
  );
}
