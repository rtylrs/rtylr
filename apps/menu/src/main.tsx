import { Suspense } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/components';
import { store, persistor } from './redux/store';
import { Provider } from 'react-redux';
import { I18nextProvider } from 'react-i18next';
import i18next from './services/i18next';
import { PersistGate } from 'redux-persist/integration/react';
import './globals.css';

injectTheme();

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense fallback={<FullScreenLoader />}>
      <I18nextProvider i18n={i18next}>
        <Provider store={store}>
          <PersistGate loading={null} persistor={persistor}>
            <BrowserRouter>
              <App />
            </BrowserRouter>
          </PersistGate>
        </Provider>
      </I18nextProvider>
    </Suspense>
  );
}
