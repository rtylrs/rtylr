import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { Toaster } from 'sonner';
import '@voxire/ui/toast.css';
import { CookiesProvider } from 'react-cookie';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { bootObservability } from '@voxire/observability';
import { publicRecruitRouter } from './router';

import './shared/services/i18next';
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
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <RouterProvider router={publicRecruitRouter} />
        <Toaster
          className="rtylr-toaster"
          expand={false}
          position="top-center"
          offset={20}
          mobileOffset={16}
        />
      </CookiesProvider>
    </QueryClientProvider>
  );
}
