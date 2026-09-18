import { Suspense, useEffect, useState } from 'react';
import { injectTheme } from '@voxire/config';
import ReactDOM from 'react-dom/client';
import App from './App';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import i18next from './shared/services/api/i18next';
import { Toaster } from 'sonner';
import { bootObservability } from '@voxire/observability';
import './globals.css';

const I18nGate: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const isReady = () => i18next.isInitialized && i18next.hasLoadedNamespace('server');
  const [ready, setReady] = useState(isReady);

  useEffect(() => {
    if (isReady()) {
      setReady(true);
      return undefined;
    }

    const markReady = () => {
      if (isReady()) setReady(true);
    };

    i18next.on('initialized', markReady);
    i18next.on('loaded', markReady);

    return () => {
      i18next.off('initialized', markReady);
      i18next.off('loaded', markReady);
    };
  }, []);

  if (!ready) {
    return (
      <div className="min-h-screen bg-pos-bg text-pos-text flex items-center justify-center">
        <div className="text-3xl font-bold tracking-tight">rtylr</div>
      </div>
    );
  }

  return <>{children}</>;
};

injectTheme();

bootObservability({ app: 'rtylr-pos' });

const rootElement = document.getElementById('root');
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
  },
});

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense fallback={null}>
      <I18nextProvider i18n={i18next}>
        <I18nGate>
          <QueryClientProvider client={queryClient}>
            <BrowserRouter>
              <App />
              <Toaster expand={false} richColors />
            </BrowserRouter>
          </QueryClientProvider>
        </I18nGate>
      </I18nextProvider>
    </Suspense>
  );
}
