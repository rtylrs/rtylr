// Schema.org builders for per-page JSON-LD. Each function returns the
// raw object we hand to HelmetContainer's `structuredData` prop - the
// component stringifies + embeds.
//
// Keeping these in one file means a rename of a field (e.g. the
// hiringOrganization logo URL) only happens in one place, and the
// per-page code stays to three lines.

const BASE = 'https://rtylr.com';

// Stable node identity for the publisher. _document.tsx emits the full
// Organization on every page and several pages emit their own (about,
// contact). Without an @id those are separate anonymous nodes that a
// consumer cannot tell are the same company; with it they merge.
const ORG_ID = `${BASE}/#organization`;

const ORG = {
  '@type': 'Organization',
  '@id': ORG_ID,
  name: 'rtylr',
  url: BASE,
  logo: `${BASE}/favicon/android-chrome-512x512.png`,
  sameAs: ['https://voxire.com'],
} as const;

interface BreadcrumbInput {
  label: string;
  path: string;
}

/**
 * breadcrumbList builds a schema.org BreadcrumbList. The last element
 * should be the current page. `path` is the URL path *with* trailing
 * slash so it matches the canonical.
 */
export function breadcrumbList(items: BreadcrumbInput[]): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((it, idx) => ({
      '@type': 'ListItem',
      position: idx + 1,
      name: it.label,
      item: `${BASE}${it.path}`,
    })),
  };
}

/**
 * productSchema describes one rtylr app (POS, Flow, HR, etc.) as a
 * SoftwareApplication. Google surfaces these in rich-result carousels;
 * the per-app description + category lets each app rank for its own
 * queries instead of always deferring to the site-wide schema.
 */
export function productSchema(opts: {
  name: string;
  path: string;
  description: string;
  category?: string;
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: opts.name,
    url: `${BASE}${opts.path}`,
    description: opts.description,
    applicationCategory: opts.category ?? 'BusinessApplication',
    operatingSystem: 'Web, iOS, Android',
    publisher: ORG,
    offers: {
      '@type': 'Offer',
      price: '0',
      priceCurrency: 'USD',
      description: 'Free tier available',
    },
  };
}

/**
 * organizationContact is the ContactPage + Organization block for the
 * /contact page. Gives Google an explicit signal that this page is
 * the canonical contact entry point, with phone / email / hours.
 */
export function organizationContactSchema(): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'ContactPage',
    url: `${BASE}/contact/`,
    mainEntity: {
      ...ORG,
      contactPoint: [
        {
          '@type': 'ContactPoint',
          contactType: 'customer support',
          email: 'info@voxire.com',
          telephone: '+961-3-940-708',
          availableLanguage: ['English', 'Arabic', 'French'],
        },
      ],
    },
  };
}

/**
 * Compound: emit both a breadcrumb and a product schema together.
 * next/head doesn't support multiple JSON-LD blocks from a single
 * structuredData prop, so we wrap both in a Graph container.
 */
export function productWithBreadcrumbs(opts: {
  name: string;
  path: string;
  description: string;
  category?: string;
  faqs?: FAQEntry[];
}): Record<string, unknown> {
  const graph: Record<string, unknown>[] = [
    productSchema(opts),
    breadcrumbList([
      { label: 'rtylr', path: '/' },
      { label: opts.name, path: opts.path },
    ]),
  ];
  if (opts.faqs && opts.faqs.length > 0) {
    graph.push(faqPageSchemaRaw(opts.faqs));
  }
  return {
    '@context': 'https://schema.org',
    '@graph': graph,
  };
}

export interface FAQEntry {
  question: string;
  answer: string;
}

/**
 * faqPageSchema wraps an FAQEntry[] in the outer @context so it can
 * be passed to HelmetContainer standalone. Use faqPageSchemaRaw when
 * composing into a @graph (it omits the outer @context to avoid
 * nesting warnings in Google's Rich Results tester).
 *
 * Why FAQ schema matters for AI SEO: ChatGPT / Perplexity / AI
 * Overviews all preferentially extract from FAQPage-marked sections
 * because the Q/A shape matches what users typed. Per Princeton's
 * GEO research, structured FAQ markup is one of the highest-ROI
 * optimizations you can make.
 */
export function faqPageSchema(faqs: FAQEntry[]): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    ...faqPageSchemaRaw(faqs),
  };
}

function faqPageSchemaRaw(faqs: FAQEntry[]): Record<string, unknown> {
  return {
    '@type': 'FAQPage',
    mainEntity: faqs.map((f) => ({
      '@type': 'Question',
      name: f.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: f.answer,
      },
    })),
  };
}

/**
 * pricingPageSchema bundles the product + an AggregateOffer
 * (listing each tier) + the visible FAQ into a single @graph.
 * Google's rich-results crawler uses AggregateOffer to render
 * price ranges in the SERP, and the FAQ feeds AI Overview snippets.
 *
 * Prices arrive as strings in the plan's currency so we can express
 * "0" (Free) and "Custom" (Enterprise) through the same shape without
 * forcing callers to branch.
 */
interface PricingTier {
  name: string;
  price: string;
  priceCurrency?: string;
  description: string;
  url?: string;
}

export function pricingPageSchema(opts: {
  tiers: PricingTier[];
  faqs: FAQEntry[];
}): Record<string, unknown> {
  const numericPrices = opts.tiers.map((t) => Number(t.price)).filter((n) => Number.isFinite(n));
  const high = numericPrices.length > 0 ? Math.max(...numericPrices) : 0;

  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'Product',
        name: 'rtylr',
        description: 'All-in-one business operating system',
        brand: { '@type': 'Brand', name: 'rtylr' },
        offers: {
          '@type': 'AggregateOffer',
          priceCurrency: 'USD',
          lowPrice: '0',
          highPrice: String(high),
          offerCount: String(opts.tiers.length),
          offers: opts.tiers.map((t) => {
            const num = Number(t.price);
            return {
              '@type': 'Offer',
              name: t.name,
              description: t.description,
              price: Number.isFinite(num) ? String(num) : t.price,
              priceCurrency: t.priceCurrency ?? 'USD',
              ...(t.url ? { url: t.url } : {}),
            };
          }),
        },
      },
      faqPageSchemaRaw(opts.faqs),
    ],
  };
}

export function organizationSchema(): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    '@id': ORG_ID,
    name: 'rtylr',
    url: BASE,
    logo: `${BASE}/favicon/android-chrome-512x512.png`,
    description:
      'All-in-one business OS for modern teams. POS, ERP, CRM, HR, Finance, and Flow in a single platform.',
    foundingDate: '2024',
    sameAs: ['https://voxire.com'],
    address: {
      '@type': 'PostalAddress',
      addressCountry: 'LB',
      addressLocality: 'Beirut',
    },
    contactPoint: {
      '@type': 'ContactPoint',
      contactType: 'customer support',
      email: 'info@voxire.com',
      telephone: '+961-3-940-708',
      availableLanguage: ['English', 'Arabic', 'French'],
    },
  };
}

export function webSiteSchema(): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    name: 'rtylr',
    url: BASE,
    description: 'Your entire business, one platform.',
    publisher: { '@type': 'Organization', name: 'rtylr', url: BASE },
    potentialAction: {
      '@type': 'SearchAction',
      target: `${BASE}/search?q={search_term_string}`,
      'query-input': 'required name=search_term_string',
    },
  };
}

export function webPageSchema(opts: {
  name: string;
  path: string;
  description: string;
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'WebPage',
        name: opts.name,
        url: `${BASE}${opts.path}`,
        description: opts.description,
        isPartOf: { '@type': 'WebSite', name: 'rtylr', url: BASE },
        publisher: ORG,
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: opts.name, path: opts.path },
      ]),
    ],
  };
}

export function blogPostingSchema(opts: {
  title: string;
  slug: string;
  description: string;
  datePublished: string;
  dateModified?: string;
}): Record<string, unknown> {
  const url = `${BASE}/blog/${opts.slug}/`;
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'BlogPosting',
        headline: opts.title,
        description: opts.description,
        url,
        mainEntityOfPage: { '@type': 'WebPage', '@id': url },
        datePublished: opts.datePublished,
        dateModified: opts.dateModified ?? opts.datePublished,
        author: ORG,
        publisher: ORG,
        image: `${BASE}/og-default.png`,
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: 'Blog', path: '/blog/' },
        { label: opts.title, path: `/blog/${opts.slug}/` },
      ]),
    ],
  };
}
