import { Html, Head, Main, NextScript } from 'next/document';
import type { DocumentContext, DocumentInitialProps } from 'next/document';

import { CONSENT_MODE_BOOTSTRAP } from '../lib/consentMode';
import { localeFromPath, type Locale } from '../lib/locales';
import { organizationSchema } from '../lib/structured-data';

interface DocProps extends DocumentInitialProps {
  htmlLang: string;
  dir: Locale['dir'];
}

/**
 * lang and dir come from the route, not from a cookie or from storage.
 *
 * This used to be a hardcoded lang="en" with no dir at all. Deriving it
 * here rather than in an effect matters because the value has to be in the
 * markup the export produces: a screen reader, a browser choosing a font,
 * and a hyphenation engine all read the attribute before any JavaScript
 * runs, and an RTL page that paints LTR first visibly reflows.
 *
 * Reading the route rather than a stored preference is deliberate. A stored
 * value disagrees with the URL the moment someone shares a link, and the
 * URL is the thing that was requested.
 *
 * Worth being accurate about the payoff: Google states it does not use
 * lang or hreflang to detect page language, so this is an accessibility
 * and correctness fix, not a ranking one.
 */
export default function Document({ htmlLang, dir }: DocProps) {
  return (
    <Html lang={htmlLang} dir={dir}>
      <Head>
        <meta charSet="UTF-8" />

        {/* Basic consent mode loads Google Analytics only after opt-in. */}
        <script dangerouslySetInnerHTML={{ __html: CONSENT_MODE_BOOTSTRAP }} />

        {/* PWA manifest - installable on iOS / Android */}
        <link rel="manifest" href="/site.webmanifest" />
        <meta name="theme-color" content="#fafafa" />

        {/*
          Sitewide structured data is Organization and nothing else,
          because Organization is the only thing that is true on every
          page. This block used to emit a SoftwareApplication carrying
          an Offer of `price: '0'`, which was two separate false
          claims: /legal/privacy-policy/ is not a software application,
          and no tier on /pricing/ is asserted to be free.

          Anything product-shaped belongs on the page that renders the
          product. See lib/structured-data.ts for the per-page builders
          and the pages that call them.
        */}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(organizationSchema()) }}
        />
        {/* Favicons */}
        <link rel="icon" type="image/x-icon" href="/favicon/favicon.ico" />
        <link rel="icon" type="image/png" sizes="16x16" href="/favicon/favicon-16x16.png" />
        <link rel="icon" type="image/png" sizes="32x32" href="/favicon/favicon-32x32.png" />
        <link rel="apple-touch-icon" sizes="180x180" href="/favicon/apple-touch-icon.png" />
        <link
          rel="icon"
          type="image/png"
          sizes="192x192"
          href="/favicon/android-chrome-192x192.png"
        />
        <link
          rel="icon"
          type="image/png"
          sizes="512x512"
          href="/favicon/android-chrome-512x512.png"
        />

        {/* Inter fallback font - DM Sans itself is self-hosted by next/font */}
        <link
          rel="preload"
          href="/fonts/Inter-Regular.ttf"
          as="font"
          type="font/ttf"
          crossOrigin=""
        />
      </Head>
      <body>
        <Main />
        <NextScript />
      </body>
    </Html>
  );
}

Document.getInitialProps = async (ctx: DocumentContext): Promise<DocProps> => {
  const initial = await ctx.defaultGetInitialProps(ctx);
  // ctx.pathname is the route pattern; ctx.asPath is the resolved URL. For a
  // static export the two agree except on dynamic routes, where asPath is
  // the one carrying the locale segment.
  const locale = localeFromPath(ctx.asPath ?? ctx.pathname ?? '/');
  return { ...initial, htmlLang: locale.htmlLang, dir: locale.dir };
};
