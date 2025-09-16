import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/shared';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import i18next from './services/i18next';
import { createTheme } from '@mui/material/styles';
import { ThemeProvider } from '@mui/material/styles';
import { Toaster } from 'sonner';
import Modal from 'react-modal';
import { queryClient } from './lib/queryClient';

import './index.scss';

import 'react-lazy-load-image-component/src/effects/blur.css';
import '@voxire/shared/styles';

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
      <Suspense fallback={<FullScreenLoader />}>
        <I18nextProvider i18n={i18next}>
          <QueryClientProvider client={queryClient}>
            <BrowserRouter>
              <App />
              <Toaster expand={false} richColors />
            </BrowserRouter>
          </QueryClientProvider>
        </I18nextProvider>
      </Suspense>
    </ThemeProvider>
  );
} else {
  console.error('Root element not found');
  document.body.innerHTML = '<div>The App Failed To Load</div>';
}
