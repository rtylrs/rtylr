import { Suspense } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import { bootObservability } from '@voxire/observability';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/i18next';
import './globals.css';

injectTheme();

bootObservability({ app: 'menu' });

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense fallback={<FullScreenLoader />}>
      <I18nextProvider i18n={i18next}>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </I18nextProvider>
    </Suspense>
  );
}
