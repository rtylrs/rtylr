import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { Toaster } from 'sonner';
import '@voxire/ui/toast.css';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { HelmetProvider } from 'react-helmet-async';
import { bootObservability } from '@voxire/observability';
import { publicRecruitRouter } from './router';
import i18next from './shared/services/i18next';
import './globals.css';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: 1, staleTime: 30_000 },
  },
});

bootObservability({ app: 'recruit' });

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <HelmetProvider>
      <QueryClientProvider client={queryClient}>
        <CookiesProvider>
          <I18nextProvider i18n={i18next}>
            <RouterProvider router={publicRecruitRouter} />
            <Toaster
              className="rtylr-toaster"
              expand={false}
              position="top-center"
              offset={20}
              mobileOffset={16}
            />
          </I18nextProvider>
        </CookiesProvider>
      </QueryClientProvider>
    </HelmetProvider>
  );
}
