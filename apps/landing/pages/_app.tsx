import { useEffect, useState } from 'react';
import { injectTheme } from '@voxire/config';
import type { AppProps } from 'next/app';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { CookiesProvider } from 'react-cookie';

import i18next from '../services/i18next';

// Import global styles
import '@voxire/components/styles';
import '../styles/index.scss';


function MyApp({ Component, pageProps }: AppProps) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            retry: 1,
            refetchOnWindowFocus: false,
          },
          mutations: {
            retry: 1,
          },
        },
      })
  );

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const Modal = require('react-modal');
      Modal.setAppElement('#__next');
    }
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
            <I18nextProvider i18n={i18next}>
            <Component {...pageProps} />
          </I18nextProvider>
        </CookiesProvider>
    </QueryClientProvider>
  );
}

export default MyApp;
