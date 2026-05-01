import { Suspense } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import { bootObservability } from '@voxire/observability';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/i18next';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import './globals.css';

const queryClient = new QueryClient();

injectTheme();

bootObservability({ app: 'menu' });

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense fallback={<FullScreenLoader />}>
      <QueryClientProvider client={queryClient}>
        <I18nextProvider i18n={i18next}>
          <BrowserRouter>
            <App />
          </BrowserRouter>
        </I18nextProvider>
      </QueryClientProvider>
    </Suspense>
  );
}
