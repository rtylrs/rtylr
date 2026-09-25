import Head from 'next/head';
import { useRouter } from 'next/router';

import { DEFAULT_LOCALE, findLocale, localeFromPath, localisePath } from '../lib/locales';

interface HelmetContainerProps {
  title?: string;
  description?: string;
  image?: string;
  url?: string;
  author?: string;
  keywords?: string;
  robots?: string;
  ogType?: string;
  ogLocale?: string;
  ogSiteName?: string;
  ogDescription?: string;
  twitterCard?: string;
  twitterSite?: string;
  twitterCreator?: string;
  customMetaTags?: Record<string, string>[];
  favicon?: string;
  // Canonical URL for this page. Absolute (https://rtylr.com/path/)
  // so Google doesn't resolve it against whatever host the HTML was
  // delivered from. Emits both <link rel="canonical"> and og:url so
  // the social-preview URL and the search-indexed URL stay aligned.
  canonical?: string;
  // JSON-LD structured data. Pass the object; we stringify. Sits
  // alongside the sitewide Organization block in _document.tsx, which
  // is the only schema that is true on every page. Anything
  // product-shaped belongs here, on the page that renders it.
  structuredData?: Record<string, unknown>;
}

// Defaults applied when a page doesn't pass explicit props. Centralised
// so missing metadata on any page still produces a coherent social
// card and a valid twitter preview - the audit finding was that most
// pages had title+description but no OG image or Twitter card, which
// made every shared link render as a text-only preview.
const OG_BASE = 'https://rtylr.com';
const DEFAULT_OG_SITE_NAME = 'rtylr';
const DEFAULT_OG_LOCALE = 'en_US';
const DEFAULT_OG_TYPE = 'website';
const DEFAULT_TWITTER_CARD = 'summary_large_image';

const DEFAULT_OG_IMAGE = `${OG_BASE}/og-default.png`;

/**
 * Map a canonical URL to its per-page OG image.
 *
 * Every page used to share og-default.png, which made a shared link to
 * /pricing/ and a shared link to /legal/terms/ look identical in a Slack
 * unfurl. The images are generated offline by
 * scripts/generate-og-images.mjs and committed under public/og.
 *
 * Derived here rather than passed per page, because a per-page prop is a
 * thing 24 pages have to remember and one will forget. The cost is that
 * this slug logic is duplicated from the generator, so
 * tests/unit/page-structure.test.mjs asserts every og:image a page
 * declares actually exists as a file in the build output. Get that
 * mapping wrong and a test fails rather than a share silently rendering
 * a broken preview.
 */
function ogImageFor(canonicalUrl: string): string {
  if (!canonicalUrl.startsWith(OG_BASE)) return DEFAULT_OG_IMAGE;
  let segments = canonicalUrl.slice(OG_BASE.length).split('/').filter(Boolean);
  // Strip a locale prefix before deriving the slug. The OG images are
  // generated once per route, not once per route per locale, so /ar/pricing/
  // shares pricing.png. Without this it would ask for ar-pricing.png and get
  // a 404 on every shared Arabic link.
  if (segments.length > 0 && findLocale(segments[0]) && segments[0] !== DEFAULT_LOCALE) {
    segments = segments.slice(1);
  }
  const slug = segments.length > 0 ? segments.join('-') : 'home';
  return `${OG_BASE}/og/${slug}.png`;
}

function HelmetContainer({
  title = 'rtylr · One suite for your entire business.',
  description = '',
  image = '',
  url = '',
  author = '',
  keywords = '',
  robots = '',
  ogType = '',
  ogLocale = '',
  ogSiteName = '',
  ogDescription = '',
  twitterCard = '',
  twitterSite = '',
  twitterCreator = '',
  customMetaTags = [],
  favicon = '',
  canonical = '',
  structuredData,
}: HelmetContainerProps) {
  // The locale of the route being rendered. Everything URL-shaped and
  // language-shaped below follows it.
  const router = useRouter();
  const routeLocale = localeFromPath(router?.asPath ?? '/');

  const effectiveOgType = ogType || DEFAULT_OG_TYPE;
  // og:locale follows the route too. It was a hardcoded en_US default, which
  // would have told every scraper that the Arabic page was American English.
  const effectiveOgLocale = ogLocale || routeLocale.ogLocale || DEFAULT_OG_LOCALE;
  const effectiveOgSiteName = ogSiteName || DEFAULT_OG_SITE_NAME;
  const effectiveTwitterCard = twitterCard || DEFAULT_TWITTER_CARD;
  // If a page passes `canonical` we use that; otherwise we fall back
  // to `url`. Both callers are common across the codebase - the
  // canonical prop is the new name we're threading in; url is the
  // legacy name most pages already pass.
  // The canonical has to be the localised URL, not the English one. Pages
  // pass a bare path like https://rtylr.com/pos/, which is correct on /pos/
  // and wrong on /ar/pos/: two URLs claiming the same canonical is how a
  // locale route gets dropped from the index in favour of the other
  // language. Derive the locale from the route and localise the path.
  const baseCanonical = canonical || url;
  const effectiveCanonical =
    baseCanonical && baseCanonical.startsWith(OG_BASE)
      ? OG_BASE + localisePath(baseCanonical.slice(OG_BASE.length) || '/', routeLocale.code)
      : baseCanonical;
  // An explicit `image` still wins, for a page that wants a bespoke card.
  // Otherwise derive it from the canonical, and only fall back to the
  // shared default when there is no canonical at all (404).
  const effectiveImage =
    image || (effectiveCanonical ? ogImageFor(effectiveCanonical) : DEFAULT_OG_IMAGE);

  return (
    <Head>
      <title>{title}</title>
      {favicon && <link rel="icon" href={favicon} type="image/x-icon" />}
      <meta name="description" content={description} />
      {author && <meta name="author" content={author} />}
      {keywords && <meta name="keywords" content={keywords} />}
      {robots && <meta name="robots" content={robots} />}
      {effectiveCanonical && <link rel="canonical" href={effectiveCanonical} />}

      <meta property="og:title" content={title} />
      <meta property="og:image" content={effectiveImage} />
      <meta property="og:image:width" content="1200" />
      <meta property="og:image:height" content="630" />
      {effectiveCanonical && <meta property="og:url" content={effectiveCanonical} />}
      <meta property="og:type" content={effectiveOgType} />
      <meta property="og:locale" content={effectiveOgLocale} />
      <meta property="og:site_name" content={effectiveOgSiteName} />
      {(ogDescription || description) && (
        <meta property="og:description" content={ogDescription || description} />
      )}

      <meta name="twitter:card" content={effectiveTwitterCard} />
      <meta name="twitter:title" content={title} />
      {description && <meta name="twitter:description" content={description} />}
      <meta name="twitter:image" content={effectiveImage} />
      {twitterSite && <meta name="twitter:site" content={twitterSite} />}
      {twitterCreator && <meta name="twitter:creator" content={twitterCreator} />}

      {customMetaTags && customMetaTags.map((tag, index) => <meta key={index} {...tag} />)}

      {structuredData && (
        <script
          type="application/ld+json"
          // next/head requires the raw string form; React won't run
          // our JSON-LD, it just embeds it for crawlers.
          dangerouslySetInnerHTML={{ __html: JSON.stringify(structuredData) }}
        />
      )}
    </Head>
  );
}

export default HelmetContainer;
