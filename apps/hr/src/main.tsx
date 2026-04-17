import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { Toaster } from 'sonner';
import App from './App';
import i18next from './shared/services/i18next';
import { queryClient } from './shared/lib/react-query';
import './globals.css';

const stored = localStorage.getItem('hr-theme') || 'light';
const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
const resolved = stored === 'system' ? (systemDark ? 'dark' : 'light') : stored;
document.documentElement.setAttribute('data-theme', resolved);

function FullScreenLoader(): React.JSX.Element {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      style={{ background: '#0a0a0a' }}
    >
      <div
        className="h-8 w-8 rounded-full border-2 animate-spin"
        style={{ borderColor: 'rgba(139,92,246,0.2)', borderTopColor: '#8b5cf6' }}
      />
    </div>
  );
}

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <QueryClientProvider client={queryClient}>
      <I18nextProvider i18n={i18next}>
        <BrowserRouter>
          <Suspense fallback={<FullScreenLoader />}>
            <App />
            <Toaster expand={false} richColors position="top-right" />
          </Suspense>
        </BrowserRouter>
      </I18nextProvider>
    </QueryClientProvider>
  );
}
