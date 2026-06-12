import { Suspense } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/api/i18next';
import { Toaster } from 'sonner';
import { bootObservability } from '@voxire/observability';
import './globals.css';

injectTheme();

bootObservability({ app: 'vbitez' });

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense fallback={null}>
      <I18nextProvider i18n={i18next}>
        <BrowserRouter>
          <App />
          <Toaster expand={false} richColors />
        </BrowserRouter>
      </I18nextProvider>
    </Suspense>
  );
}
