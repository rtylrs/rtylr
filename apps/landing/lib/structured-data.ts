// Schema.org builders for per-page JSON-LD. Each function returns the
// raw object we hand to HelmetContainer's `structuredData` prop — the
// component stringifies + embeds.
//
// Keeping these in one file means a rename of a field (e.g. the
// hiringOrganization logo URL) only happens in one place, and the
// per-page code stays to three lines.

const BASE = 'https://rtylr.com';
const ORG = {
  '@type': 'Organization',
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
