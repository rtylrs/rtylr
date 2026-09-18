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
 *
 * No `offers`. A price is a fact, and rtylr does not publish one yet:
 * these pages show no price at all, so asserting `price: '0'` on each
 * of them was a claim about a paid product that no page backed up. It
 * comes back when a real number ships on /pricing/, generated from the
 * same data that page renders.
 */
export function productSchema(opts: {
  name: string;
  path: string;
  description: string;
  category?: string;
  subCategory?: string;
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: opts.name,
    url: `${BASE}${opts.path}`,
    description: opts.description,
    applicationCategory: opts.category ?? 'BusinessApplication',
    // Every module used to declare the same applicationCategory, which
    // told a consumer that /pos/ and /hr/ were the same kind of thing.
    // The sub-category is what distinguishes them.
    ...(opts.subCategory ? { applicationSubCategory: opts.subCategory } : {}),
    operatingSystem: 'Web, iOS, Android',
    publisher: ORG,
  };
}

/**
 * productPageSchema describes a physical product line, which is what
 * /hardware/ is: registers, printers, scanners and drawers. It was
 * emitting WebPage, which is true of any URL and says nothing.
 *
 * No `offers`, for the same reason nothing else has one: rtylr does not
 * publish hardware prices on the page.
 */
export function productPageSchema(opts: {
  name: string;
  path: string;
  description: string;
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'Product',
        name: opts.name,
        url: `${BASE}${opts.path}`,
        description: opts.description,
        brand: { '@type': 'Brand', name: 'rtylr' },
        manufacturer: ORG,
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: opts.name, path: opts.path },
      ]),
    ],
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
  subCategory?: string;
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      productSchema(opts),
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: opts.name, path: opts.path },
      ]),
    ],
  };
}

export interface FAQEntry {
  question: string;
  answer: string;
}

/**
 * pricingPageSchema describes /pricing/ as a Product plus a breadcrumb.
 *
 * There is deliberately no Offer and no AggregateOffer. It used to emit
 * both, built from placeholder numbers, which meant the page asserted a
 * `lowPrice` of 0 and a `highPrice` derived from figures that were never
 * real. A price is a fact. rtylr does not publish one yet, and /pricing/
 * now shows tier names and inclusions only, so there is nothing to
 * express and nothing to get wrong.
 *
 * When real pricing ships, generate the Offer nodes from the same tier
 * data the page renders, in one pass, so the markup cannot drift from
 * the visible text.
 *
 * There is also no FAQPage: see the note on faqs.ts. The questions stay
 * on the page, where a reader and an extractor can both find them.
 */
export function pricingPageSchema(): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'Product',
        name: 'rtylr',
        description: 'All-in-one business operating system',
        brand: { '@type': 'Brand', name: 'rtylr' },
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: 'Pricing', path: '/pricing/' },
      ]),
    ],
  };
}

/**
 * itemListSchema is plain semantic markup for a page that is genuinely a
 * list, which /integrations/ is.
 *
 * Expect no rich result from it. Google's Carousel requires ItemList be
 * combined with Recipe, Course, Restaurant or Movie, and an integrations
 * directory is none of those. It is here because it describes the page
 * accurately for anything reading the graph, not because it earns a SERP
 * feature. Do not let anyone sell it as one.
 */
export function itemListSchema(opts: {
  name: string;
  path: string;
  description: string;
  items: string[];
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'CollectionPage',
        name: opts.name,
        url: `${BASE}${opts.path}`,
        description: opts.description,
        isPartOf: { '@type': 'WebSite', name: 'rtylr', url: BASE },
        publisher: ORG,
        mainEntity: {
          '@type': 'ItemList',
          numberOfItems: opts.items.length,
          itemListElement: opts.items.map((name, idx) => ({
            '@type': 'ListItem',
            position: idx + 1,
            name,
          })),
        },
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: opts.name, path: opts.path },
      ]),
    ],
  };
}

/**
 * aboutPageSchema for /about/. AboutPage is the specific type for a page
 * describing the organisation, and naming the Organization by @id points
 * at the same node _document.tsx emits rather than restating it.
 */
export function aboutPageSchema(opts: { description: string }): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'AboutPage',
        name: 'About rtylr',
        url: `${BASE}/about/`,
        description: opts.description,
        mainEntity: { '@id': ORG_ID },
        isPartOf: { '@type': 'WebSite', name: 'rtylr', url: BASE },
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: 'About', path: '/about/' },
      ]),
    ],
  };
}

/**
 * contactPageSchema for /demo/, which is a contact form for a sales
 * conversation and was declaring itself a generic WebPage.
 */
export function contactPageSchema(opts: {
  name: string;
  path: string;
  description: string;
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'ContactPage',
        name: opts.name,
        url: `${BASE}${opts.path}`,
        description: opts.description,
        mainEntity: { '@id': ORG_ID },
        isPartOf: { '@type': 'WebSite', name: 'rtylr', url: BASE },
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: opts.name, path: opts.path },
      ]),
    ],
  };
}

/**
 * blogIndexSchema for /blog/: a Blog whose posts are enumerated, so a
 * consumer reading the graph can see what is on the page without
 * parsing the markup.
 */
export function blogIndexSchema(opts: {
  description: string;
  posts: { title: string; slug: string; datePublished: string }[];
}): Record<string, unknown> {
  return {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'Blog',
        name: 'rtylr blog',
        url: `${BASE}/blog/`,
        description: opts.description,
        publisher: ORG,
        blogPost: opts.posts.map((p) => ({
          '@type': 'BlogPosting',
          headline: p.title,
          url: `${BASE}/blog/${p.slug}/`,
          datePublished: p.datePublished,
        })),
      },
      breadcrumbList([
        { label: 'rtylr', path: '/' },
        { label: 'Blog', path: '/blog/' },
      ]),
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
