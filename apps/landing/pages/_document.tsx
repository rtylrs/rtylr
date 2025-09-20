import { Html, Head, Main, NextScript } from 'next/document';

export default function Document() {
  return (
    <Html lang="en">
      <Head>
        <meta charSet="UTF-8" />
        <link rel="icon" type="image/x-icon" href="/favicon.ico" />

        {/* Fonts */}
        <link rel="preload" href="/fonts/ggsans-Normal.ttf" as="font" type="font/ttf" crossOrigin="" />
        <link rel="preload" href="/fonts/ggsans-Bold.ttf" as="font" type="font/ttf" crossOrigin="" />
        <link rel="preload" href="/fonts/PlusJakartaSans.ttf" as="font" type="font/ttf" crossOrigin="" />
        <link rel="preload" href="/fonts/Inter-Regular.ttf" as="font" type="font/ttf" crossOrigin="" />
        <link rel="preload" href="/fonts/PlusJakartaSans-ExtraBold.ttf" as="font" type="font/ttf" crossOrigin="" />
      </Head>
      <body>
        <Main />
        <NextScript />
      </body>
    </Html>
  );
}