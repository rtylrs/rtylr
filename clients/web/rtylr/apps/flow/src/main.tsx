import React, { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider } from '@tanstack/react-router';
import { Toaster } from 'sonner';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { bootObservability } from '@voxire/observability';
import { loadRuntimeConfig } from '@voxire/config';
import { initAppearance } from '@voxire/ui';
import { APPEARANCE_DEFAULTS, LEGACY_THEME_STORAGE_KEY } from '@voxire/utils';
import { router } from './router';
import { FLOW_APPEARANCE_STORAGE_KEY } from './shared/stores/appearanceStore';
import { ACCENT_DEFAULT } from './shared/lib/colorUtils';
import i18n from './shared/services/i18next';
import './globals.css';

// Apply the persisted appearance before the first render so the first paint is
// already the right palette rather than the stylesheet default.
//
// This is the pre-paint half of a pair. AppearanceProvider, mounted in
// FlowLayout, owns every write from mount onwards; this call covers the window
// before it exists, which is the AppGuard fetch. The workspace accent is not
// known yet at this point, so the brand teal stands in until the layout
// resolves the active workspace, exactly as it did before.
initAppearance({
  storageKey: FLOW_APPEARANCE_STORAGE_KEY,
  defaults: APPEARANCE_DEFAULTS,
  workspaceAccent: ACCENT_DEFAULT,
  legacyModeKey: LEGACY_THEME_STORAGE_KEY,
});

bootObservability({ app: 'flow' });

const BootLoader = (): React.JSX.Element => {
  return (
    <div
      style={{
        display: 'flex',
        height: '100vh',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--flow-bg)',
        color: 'var(--flow-text-muted)',
        fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
        fontSize: 'var(--text-section)',
      }}
    >
      {i18n.isInitialized ? i18n.t('flow.shell.loadingApp') : 'Loading…'}
    </div>
  );
};

const i18nReady = i18n.isInitialized
  ? Promise.resolve()
  : new Promise<void>((resolve) => {
      i18n.on('initialized', () => resolve());
    });

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60,
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const rootElement = document.getElementById('root');

if (rootElement) {
  const root = ReactDOM.createRoot(rootElement);
  // Await runtime config (self-host) before rendering; loader meanwhile, static
  // hosted defaults on failure. Hosted is unchanged.
  root.render(
    <I18nextProvider i18n={i18n}>
      <BootLoader />
    </I18nextProvider>
  );
  void Promise.all([loadRuntimeConfig(), i18nReady]).finally(() => {
    root.render(
      <I18nextProvider i18n={i18n}>
        <QueryClientProvider client={queryClient}>
          <Suspense fallback={<BootLoader />}>
            <RouterProvider router={router} />
            <Toaster expand={false} richColors />
          </Suspense>
        </QueryClientProvider>
      </I18nextProvider>
    );
  });
}
