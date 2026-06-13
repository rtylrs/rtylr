import { Suspense } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/api/i18next';
import { Toaster } from 'sonner';
import { bootObservability } from '@voxire/observability';
import './globals.css';

injectTheme();

bootObservability({ app: 'rtylr-pos' });

const rootElement = document.getElementById('root');
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
  },
});

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense fallback={null}>
      <I18nextProvider i18n={i18next}>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <App />
            <Toaster expand={false} richColors />
          </BrowserRouter>
        </QueryClientProvider>
      </I18nextProvider>
    </Suspense>
  );
}
