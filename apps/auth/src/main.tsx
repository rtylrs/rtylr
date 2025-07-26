import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import Loader from './layouts/loader/Loader';
import { store, persistor } from './redux/store';
import { Provider } from 'react-redux';
import { I18nextProvider } from 'react-i18next';
import i18next from './services/i18next';
import { createTheme } from '@mui/material/styles';
import { ThemeProvider } from '@mui/material/styles';
import { Toaster } from 'sonner';
import { PersistGate } from 'redux-persist/integration/react';
import Modal from 'react-modal';

import 'react-lazy-load-image-component/src/effects/blur.css';

const theme = createTheme({
  palette: {
    primary: {
      main: '#040428',
    },
    secondary: {
      main: '#4485c4',
    },
  },
});

Modal.setAppElement('#root');

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <ThemeProvider theme={theme}>
      <Suspense fallback={<Loader />}>
        <I18nextProvider i18n={i18next}>
          <Provider store={store}>
            <PersistGate loading={null} persistor={persistor}>
              <BrowserRouter>
                <App />
                <Toaster expand={false} richColors />
              </BrowserRouter>
            </PersistGate>
          </Provider>
        </I18nextProvider>
      </Suspense>
    </ThemeProvider>
  );
} else {
  console.error('Root element not found');
  document.body.innerHTML = '<div>The App Failed To Load</div>';
}
