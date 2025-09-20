import { useEffect, useState } from 'react';
import type { AppProps } from 'next/app';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { I18nextProvider } from 'react-i18next';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { CookiesProvider } from 'react-cookie';

import i18next from '../src/services/i18next';

// Import global styles
import '@voxire/shared/styles';
import '../src/styles/index.scss';

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

function MyApp({ Component, pageProps }: AppProps) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000, // 1 minute
            cacheTime: 5 * 60 * 1000, // 5 minutes
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
    // Set modal app element for accessibility - only on client side
    if (typeof window !== 'undefined') {
      const Modal = require('react-modal');
      Modal.setAppElement('#__next');
    }
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <ThemeProvider theme={theme}>
          <I18nextProvider i18n={i18next}>
            <Component {...pageProps} />
          </I18nextProvider>
        </ThemeProvider>
      </CookiesProvider>
      {process.env.NODE_ENV === 'development' && (
        <ReactQueryDevtools initialIsOpen={false} />
      )}
    </QueryClientProvider>
  );
}

export default MyApp;