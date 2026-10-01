import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { Toaster } from 'sonner';
import { bootObservability } from '@voxire/observability';
import { router } from './router';
import i18next from './shared/services/i18next';
import { queryClient } from './shared/lib/react-query';
import './globals.css';

bootObservability({ app: 'hr' });

const stored = localStorage.getItem('hr-theme') || 'light';
const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
const resolved = stored === 'system' ? (systemDark ? 'dark' : 'light') : stored;
document.documentElement.setAttribute('data-theme', resolved);

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <QueryClientProvider client={queryClient}>
      <I18nextProvider i18n={i18next}>
        <RouterProvider router={router} />
        <Toaster expand={false} richColors position="top-right" />
      </I18nextProvider>
    </QueryClientProvider>
  );
}
