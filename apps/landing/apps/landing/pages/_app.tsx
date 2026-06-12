import { useEffect, useState } from 'react';
import { injectTheme } from '@voxire/config';
import { bootObservability } from '@voxire/observability';
import { CookieBanner } from '@voxire/consent';
import type { AppProps } from 'next/app';
import { MotionConfig } from 'motion/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { CookiesProvider } from 'react-cookie';

import i18next from '../services/i18next';
import { initFeatureVariant } from '../lib/featureVariant';

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

  // Next.js SSG: boot observability client-side only. Sentry init
  // reads from `window`, and PostHog must wait for a consent cookie
  // that only exists in the browser.
  useEffect(() => {
    bootObservability({ app: 'landing' });
  }, []);

  // Resolve the landing A/B variant and write the rtylr_variant cookie so the
  // Cloudflare worker can route returning visitors. Fails safe to variant 'a'.
  useEffect(() => {
    void initFeatureVariant();
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <I18nextProvider i18n={i18next}>
          <MotionConfig reducedMotion="user">
            <Component {...pageProps} />
          </MotionConfig>
          {/* Strict opt-in banner. Shows only until the visitor
              accepts or declines - then cookies-granted unlocks the
              PostHog init wired in bootObservability. */}
          <CookieBanner />
        </I18nextProvider>
      </CookiesProvider>
    </QueryClientProvider>
  );
}

export default MyApp;
