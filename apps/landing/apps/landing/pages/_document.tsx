import { Html, Head, Main, NextScript } from 'next/document';

import { CONSENT_MODE_BOOTSTRAP, GA_MEASUREMENT_ID } from '../lib/consentMode';
import { organizationSchema } from '../lib/structured-data';

export default function Document() {
  return (
    <Html lang="en">
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
