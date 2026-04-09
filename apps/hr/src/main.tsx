import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { Toaster } from 'sonner';
import App from './App';
import './globals.css';

// Apply persisted theme before first render to prevent flash
const stored = localStorage.getItem('hr-theme') || 'dark';
const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
const resolved = stored === 'system' ? (systemDark ? 'dark' : 'light') : stored;
document.documentElement.setAttribute('data-theme', resolved);

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      retry: 1,
    },
  },
});

function FullScreenLoader() {
  return (
    <div
      style={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: '#0a0a0a',
      }}
    >
      <div
        style={{
          width: 32,
          height: 32,
          borderRadius: '50%',
          border: '2px solid rgba(139,92,246,0.2)',
          borderTopColor: '#8b5cf6',
          animation: 'spin 0.7s linear infinite',
        }}
      />
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
    </div>
  );
}

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Suspense fallback={<FullScreenLoader />}>
          <App />
          <Toaster expand={false} richColors position="top-right" />
        </Suspense>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
