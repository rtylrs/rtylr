import Head from 'next/head';

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

function HelmetContainer({
  title = 'rtylr · Your entire business, one platform.',
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
  const effectiveImage = image || DEFAULT_OG_IMAGE;
  const effectiveOgType = ogType || DEFAULT_OG_TYPE;
  const effectiveOgLocale = ogLocale || DEFAULT_OG_LOCALE;
  const effectiveOgSiteName = ogSiteName || DEFAULT_OG_SITE_NAME;
  const effectiveTwitterCard = twitterCard || DEFAULT_TWITTER_CARD;
  // If a page passes `canonical` we use that; otherwise we fall back
  // to `url`. Both callers are common across the codebase - the
  // canonical prop is the new name we're threading in; url is the
  // legacy name most pages already pass.
  const effectiveCanonical = canonical || url;

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
