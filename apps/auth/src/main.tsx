import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { FullScreenLoader } from '@voxire/shared';
import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import i18next from './services/i18next';
import { Toaster } from 'sonner';
import Modal from 'react-modal';
import { queryClient } from './lib/queryClient';

import 'react-lazy-load-image-component/src/effects/blur.css';
import '@voxire/shared/styles';

import './index.scss';

Modal.setAppElement('#root');

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
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
  );
} else {
  console.error('Root element not found');
  document.body.innerHTML = '<div>The App Failed To Load</div>';
}
