import { useEffect, useState } from 'react';
import { injectTheme } from '@voxire/config';
import { bootObservability } from '@voxire/observability';
import { CookieBanner } from '@voxire/consent';
import type { AppProps } from 'next/app';
import { useRouter } from 'next/router';
import { AnimatePresence, MotionConfig, motion } from 'motion/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { CookiesProvider } from 'react-cookie';
import { DM_Sans } from 'next/font/google';

import i18next from '../services/i18next';
import { initFeatureVariant } from '../lib/featureVariant';
import { watchConsentMode } from '../lib/consentMode';

// Import global styles
import '@voxire/components/styles';
import '../styles/index.scss';

// Exposes --font-dm-sans for `var(--primary-font)` in styles/_variables.scss
// (see the .className below for the plain-inheritance path).
const dmSans = DM_Sans({
  subsets: ['latin'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-dm-sans',
  display: 'swap',
});

function MyApp({ Component, pageProps }: AppProps) {
  const router = useRouter();
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

  // Mirror the cookie banner's decision into Google Consent Mode. The
  // deny-by-default half runs inline in _document's head; this half
  // carries a decision made during this page view, so accepting the
  // banner unblocks GA4 without a reload.
  useEffect(() => watchConsentMode(), []);

  return (
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <I18nextProvider i18n={i18next}>
          <div className={`${dmSans.variable} ${dmSans.className}`}>
            <MotionConfig reducedMotion="user">
              {/* Opacity-only route crossfade - no transform, so fixed headers
                  inside pages are never re-anchored mid-transition. */}
              <AnimatePresence mode="wait" initial={false}>
                <motion.div
                  key={router.route}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0, transition: { duration: 0.15, ease: 'easeIn' } }}
                  transition={{ duration: 0.35, ease: 'easeOut' }}
                >
                  <Component {...pageProps} />
                </motion.div>
              </AnimatePresence>
            </MotionConfig>
            {/* Strict opt-in banner. Shows only until the visitor
                accepts or declines - then cookies-granted unlocks the
                PostHog init wired in bootObservability. */}
            <CookieBanner />
          </div>
        </I18nextProvider>
      </CookiesProvider>
    </QueryClientProvider>
  );
}

export default MyApp;
