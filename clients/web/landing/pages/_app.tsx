import { useEffect, useLayoutEffect, useMemo, useState } from 'react';
import { injectTheme } from '@voxire/config';
import { bootObservability } from '@voxire/observability';
import { CookieBanner, privacyLinkStyle } from '@voxire/consent';
import type { AppProps } from 'next/app';
import { useRouter } from 'next/router';
import { MotionConfig } from 'motion/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider, Trans } from 'react-i18next';
import { CookiesProvider } from 'react-cookie';
import { DM_Sans, IBM_Plex_Sans_Arabic } from 'next/font/google';

import i18next from '../services/i18next';
import { localeFromPath, localisePath } from '../lib/locales';
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

// Arabic display face. DM Sans ships Latin only, so an Arabic page rendered
// in it falls back per glyph to whatever the OS has, which is why this is a
// real font dependency and not a CSS-only change.
//
// Exposes --font-plex-arabic, consumed by --rtl-font in styles/_variables.scss
// and switched on by :root[dir='rtl'] .dir-font-scope in styles/index.scss.
// Nothing here changes a Latin page: the variable is defined, and no glyph on
// those pages matches the Arabic unicode-range, so the file is never fetched.
//
// preload: false is deliberate. next/font would otherwise emit a preload link
// for the Arabic woff2 into the head of every exported page, including the
// Latin ones that will never paint a single Arabic glyph. Without the preload
// the browser still fetches it as soon as Arabic text needs it, driven by the
// unicode-range on the generated @font-face.
const plexArabic = IBM_Plex_Sans_Arabic({
  subsets: ['arabic'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-plex-arabic',
  display: 'swap',
  preload: false,
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

  /**
   * One i18next instance per locale, chosen from the route.
   *
   * This has to happen during render, not in an effect. The export
   * prerenders every page on the server where there is no window, so an
   * effect-driven language means the prerendered HTML is always English:
   * /ar/pricing/ shipped with lang="ar" dir="rtl" and English body copy,
   * which is precisely the doorway page the locale registry warns about.
   * A crawler, a markdown twin and a reader without JavaScript all see that
   * HTML and never the corrected client render.
   *
   * cloneInstance rather than changeLanguage on the shared instance:
   * changeLanguage mutates module-level state, and the export renders every
   * page in one process, so page order would decide which language a page
   * came out in. A clone shares the loaded resources and carries its own
   * language, so /pricing/ and /ar/pricing/ can be rendered in any order.
   */
  const locale = localeFromPath(router.asPath);
  const i18n = useMemo(
    () => (locale.code === i18next.language ? i18next : i18next.cloneInstance({ lng: locale.code })),
    [locale.code]
  );

  // _document sets lang/dir correctly for a fresh load, but that only runs
  // once per server render. Next's router does not reload the document on
  // a client-side navigation, so without this a locale switch (or any link
  // crossing between a localised and an unlocalised page) leaves the DOM's
  // lang/dir stuck on whichever locale first loaded the page - Arabic text
  // rendered inside an ltr document, or an English page fully mirrored rtl.
  // useLayoutEffect rather than useEffect so the correction lands before
  // the browser paints the new route, not one frame after.
  const isomorphicLayoutEffect = typeof window === 'undefined' ? useEffect : useLayoutEffect;
  isomorphicLayoutEffect(() => {
    document.documentElement.lang = locale.htmlLang;
    document.documentElement.dir = locale.dir;
  }, [locale.htmlLang, locale.dir]);

  return (
    <QueryClientProvider client={queryClient}>
      <CookiesProvider>
        <I18nextProvider i18n={i18n}>
          <div
            className={`${dmSans.variable} ${plexArabic.variable} ${dmSans.className} dir-font-scope`}
          >
            {/* No route-transition wrapper here anymore: an
                AnimatePresence/motion.div opacity crossfade keyed on
                router.route used to wrap this. Removed because
                framer-motion's exit animation never actually completed for
                a route change between two top-level pages (confirmed with
                router.events firing routeChangeStart/routeChangeComplete
                normally, the new page's JS chunk loading, and
                window.next.router.route updating correctly - React itself
                just never unmounted the outgoing motion.div), so the
                incoming page's Component never mounted under mode="wait",
                and both pages' DOM coexisted indefinitely under
                mode="sync" instead of the outgoing one ever being removed.
                MotionConfig has no such lifecycle - it's a plain context
                provider - so it stays; it's the only thing that makes the
                ~30 motion/react components respect prefers-reduced-motion. */}
            <MotionConfig reducedMotion="user">
              <Component {...pageProps} />
            </MotionConfig>
            {/* Strict opt-in banner. Shows only until the visitor
                accepts or declines - then cookies-granted unlocks the
                PostHog init wired in bootObservability. Copy is passed
                in rather than looked up inside the shared package - see
                the note on CookieBannerProps in @voxire/consent. */}
            <CookieBanner
              privacyUrl={localisePath('/legal/privacy-policy/', locale.code)}
              message={
                <Trans
                  i18n={i18n}
                  i18nKey="common.cookieBanner.message"
                  components={{
                    privacyLink: (
                      <a
                        href={localisePath('/legal/privacy-policy/', locale.code)}
                        style={privacyLinkStyle}
                      />
                    ),
                  }}
                />
              }
              declineLabel={i18n.t('common.cookieBanner.decline')}
              acceptLabel={i18n.t('common.cookieBanner.accept')}
              ariaLabel={i18n.t('common.cookieBanner.ariaLabel')}
              // Below the mobile nav's full-screen overlay (z-index 2000) so
              // opening the menu doesn't leave the banner floating on top of
              // it, intercepting taps meant for the menu underneath. Still
              // above the header (1000) and ordinary page content.
              zIndex={1500}
            />
          </div>
        </I18nextProvider>
      </CookiesProvider>
    </QueryClientProvider>
  );
}

export default MyApp;
