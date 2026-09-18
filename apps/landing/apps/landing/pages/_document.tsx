import { Html, Head, Main, NextScript } from 'next/document';
import type { DocumentContext, DocumentInitialProps } from 'next/document';

import { CONSENT_MODE_BOOTSTRAP, GA_MEASUREMENT_ID } from '../lib/consentMode';
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

        {/*
          GA4, measurement ID G-6XKFL1Z8HE, property "rtylr.com".
          Placed here rather than in _app so it lands in the static export's
          document head on every page, which is the same reason voxire.com
          keeps its tag in _document.

          Installed as gtag.js directly rather than through a Tag Manager
          container: this site has no other tags to manage, and it is the
          install Google recommends on the stream setup screen. If tag
          management is wanted later, swap this block for a GTM snippet and
          configure the GA4 tag inside the container instead. Do not run both.

          afterInteractive-equivalent behaviour comes from `async`. Do not
          convert this to next/script: the export needs it in the document
          head, and that is a rule voxire's CLAUDE.md records the hard way.

          The bootstrap denies every storage type, then grants analytics
          if the visitor has already accepted. lib/consentMode.ts holds
          the snippet and the runtime half of the bridge; _app.tsx
          subscribes to later changes.

          What Consent Mode actually requires is that `consent default`
          sit ahead of `config` in the dataLayer queue, and both are in
          the one inline snippet below, so that holds by construction.
          Do not rely on the JSX order of these two tags: Next hoists
          the async gtag.js tag above the inline script in the exported
          head. That is harmless, because the inline script is
          synchronous and fills the queue during parse while the async
          library is still fetching. tests/unit/consent-mode.test.mjs
          asserts the queue order that matters.
        */}
        <script dangerouslySetInnerHTML={{ __html: CONSENT_MODE_BOOTSTRAP }} />
        <script async src={`https://www.googletagmanager.com/gtag/js?id=${GA_MEASUREMENT_ID}`} />

        {/*
          Google Preferred Sources library. Renders the button in the footer
          (layouts/Main/Header is not it: see Footer.tsx). Script URL and the
          div attribute both come from the Search Central docs read
          2026-08-23; the attribute name is the entire contract, so neither
          was guessed.

          Two things about this are open, and are flagged rather than
          decided quietly here.

          1. Consent. This is a third-party Google script loading on every
             page regardless of the visitor's cookie choice. It is a
             functional control rather than analytics, and Google's docs say
             nothing about consent, but this site takes a strict opt-in
             stance and GA4 was wired to Consent Mode for exactly this
             reason.

          2. Cost. The library injects two cross-origin iframes, and they
             hold the document's load event open. Measured against the
             built site on 2026-08-23, /pos/, median of repeated runs:

               without it   loadEventEnd  117ms,  DCL 116ms, FCP 80ms
               with it      loadEventEnd  22-29s, DCL 107ms, FCP 68ms

             First paint and DOMContentLoaded are untouched, so this is not
             an LCP or INP problem, and nothing in this codebase listens for
             `load` (checked). What it does affect is the browser's loading
             indicator and any future code keyed on window.onload. It also
             made the e2e suite go from 13s to 4 minutes until the fixture
             started blocking off-origin requests.

          Both point at the same remedy if either matters: load this
          lazily, on consent or on the footer entering the viewport,
          instead of in the head. That deviates from Google's documented
          integration, which is why it was not done unilaterally.

          No integrity hash: this is a mutable versioned library that Google
          updates in place and publishes no SRI digest for, so pinning one
          would break the button on their next release. Same situation as
          gtag.js above.
        */}
        <script async src="https://news.google.com/swg/js/v1/publisher.js" />

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
