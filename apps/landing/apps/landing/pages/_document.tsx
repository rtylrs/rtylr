import { Html, Head, Main, NextScript } from 'next/document';

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
        */}
        <script async src="https://www.googletagmanager.com/gtag/js?id=G-6XKFL1Z8HE" />
        <script
          dangerouslySetInnerHTML={{
            __html: `window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());
gtag('config', 'G-6XKFL1Z8HE');`,
          }}
        />

        {/* PWA manifest - installable on iOS / Android */}
        <link rel="manifest" href="/site.webmanifest" />
        <meta name="theme-color" content="#fafafa" />

        {/* Structured Data - Organization */}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              '@context': 'https://schema.org',
              '@type': 'SoftwareApplication',
              name: 'rtylr',
              applicationCategory: 'BusinessApplication',
              operatingSystem: 'Web, Windows, macOS, Linux',
              description:
                'All-in-one business operating system for SMBs. POS, ERP, HR, CRM, Finance, and more.',
              url: 'https://rtylr.com',
              offers: {
                '@type': 'Offer',
                price: '0',
                priceCurrency: 'USD',
                description: 'Free tier available',
              },
              publisher: {
                '@type': 'Organization',
                name: 'Voxire',
                url: 'https://voxire.com',
              },
            }),
          }}
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
