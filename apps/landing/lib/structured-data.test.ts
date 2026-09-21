import { describe, expect, it } from 'vitest';

import {
  breadcrumbList,
  organizationContactSchema,
  organizationSchema,
  pricingPageSchema,
  productSchema,
  productWithBreadcrumbs,
  webPageSchema,
  webSiteSchema,
} from './structured-data';

/**
 * JSON-LD structured data is the silent half of SEO - Google, Perplexity,
 * and ChatGPT all parse these blobs but users never see them, so a broken
 * schema can go unnoticed for weeks. These tests pin the shape and the
 * field semantics so a refactor can't quietly regress the AI-SEO wiring.
 */

describe('breadcrumbList', () => {
  it('emits a BreadcrumbList with absolute item URLs', () => {
    const out = breadcrumbList([
      { label: 'rtylr', path: '/' },
      { label: 'POS', path: '/pos/' },
    ]);
    expect(out['@type']).toBe('BreadcrumbList');
    expect(out['@context']).toBe('https://schema.org');
    const list = out.itemListElement as Array<{
      '@type': string;
      position: number;
      name: string;
      item: string;
    }>;
    expect(list).toHaveLength(2);
    expect(list[0]).toMatchObject({
      '@type': 'ListItem',
      position: 1,
      name: 'rtylr',
      item: 'https://rtylr.com/',
    });
    expect(list[1]).toMatchObject({
      position: 2,
      name: 'POS',
      item: 'https://rtylr.com/pos/',
    });
  });
});

describe('productSchema', () => {
  it('emits SoftwareApplication without any offer', () => {
    const out = productSchema({
      name: 'POS',
      path: '/pos/',
      description: 'd',
    });
    expect(out['@type']).toBe('SoftwareApplication');
    expect(out.applicationCategory).toBe('BusinessApplication');
    expect(out.url).toBe('https://rtylr.com/pos/');
    // No price is published anywhere on the site, so no module page may
    // assert one. This used to emit `price: '0'` on all six.
    expect(out.offers).toBeUndefined();
  });

  it('allows overriding applicationCategory', () => {
    const out = productSchema({
      name: 'Finance',
      path: '/finance/',
      description: 'd',
      category: 'FinanceApplication',
    });
    expect(out.applicationCategory).toBe('FinanceApplication');
  });
});

describe('productWithBreadcrumbs', () => {
  it('composes a @graph with product and breadcrumb', () => {
    const out = productWithBreadcrumbs({
      name: 'Flow',
      path: '/flow/',
      description: 'd',
    });
    expect(out['@context']).toBe('https://schema.org');
    const graph = out['@graph'] as Array<Record<string, unknown>>;
    expect(graph).toHaveLength(2);
    expect(graph[0]['@type']).toBe('SoftwareApplication');
    expect(graph[1]['@type']).toBe('BreadcrumbList');
  });

  it('never appends an FAQPage', () => {
    // Google removed the FAQ rich result in May 2026 and the questions
    // were never rendered on the page, so the markup described content
    // that did not exist. The questions live in the page now.
    const out = productWithBreadcrumbs({ name: 'Flow', path: '/flow/', description: 'd' });
    const graph = out['@graph'] as Array<Record<string, unknown>>;
    expect(graph).toHaveLength(2);
    expect(JSON.stringify(out)).not.toContain('FAQPage');
  });

  it('carries the sub category through to the application node', () => {
    const out = productWithBreadcrumbs({
      name: 'Flow',
      path: '/flow/',
      description: 'd',
      subCategory: 'Project and Task Management',
    });
    const graph = out['@graph'] as Array<Record<string, unknown>>;
    expect(graph[0].applicationSubCategory).toBe('Project and Task Management');
  });
});

describe('organizationContactSchema', () => {
  it('emits a ContactPage with contactPoint', () => {
    const out = organizationContactSchema();
    expect(out['@type']).toBe('ContactPage');
    const entity = out.mainEntity as {
      '@type': string;
      contactPoint: Array<{ email: string; contactType: string }>;
    };
    expect(entity['@type']).toBe('Organization');
    expect(entity.contactPoint[0].email).toBe('info@voxire.com');
    expect(entity.contactPoint[0].contactType).toBe('customer support');
  });
});

describe('pricingPageSchema', () => {
  it('emits Product and a breadcrumb, and nothing else', () => {
    const out = pricingPageSchema();
    const graph = out['@graph'] as Array<Record<string, unknown>>;
    expect(graph).toHaveLength(2);
    expect(graph[0]['@type']).toBe('Product');
    expect(graph[1]['@type']).toBe('BreadcrumbList');
  });

  it('carries no Offer, no AggregateOffer and no FAQPage', () => {
    // /pricing/ shows tier names and inclusions only. The moment a
    // price string appears in this schema again it is a claim the page
    // does not make, so assert its absence rather than its shape.
    const serialised = JSON.stringify(pricingPageSchema());
    expect(serialised).not.toContain('Offer');
    expect(serialised).not.toContain('price');
    expect(serialised).not.toContain('FAQPage');
  });
});

describe('organizationSchema', () => {
  it('emits an Organization type', () => {
    const out = organizationSchema();
    expect(out['@type']).toBe('Organization');
    expect(out['@context']).toBe('https://schema.org');
  });

  it('includes the rtylr name and URL', () => {
    const out = organizationSchema();
    expect(out.name).toBe('rtylr');
    expect(out.url).toBe('https://rtylr.com');
  });

  it('includes a logo URL', () => {
    const out = organizationSchema() as { logo: string };
    expect(out.logo).toContain('rtylr.com');
  });

  it('includes the founding date', () => {
    const out = organizationSchema() as { foundingDate: string };
    expect(out.foundingDate).toBe('2024');
  });

  it('includes a PostalAddress', () => {
    const out = organizationSchema() as { address: { '@type': string } };
    expect(out.address['@type']).toBe('PostalAddress');
  });

  it('includes a contactPoint with an email and phone', () => {
    const out = organizationSchema() as {
      contactPoint: { '@type': string; email: string; telephone: string };
    };
    expect(out.contactPoint['@type']).toBe('ContactPoint');
    expect(out.contactPoint.email).toBeTruthy();
    expect(out.contactPoint.telephone).toBeTruthy();
  });
});

describe('webSiteSchema', () => {
  it('emits a WebSite type with correct URL', () => {
    const out = webSiteSchema();
    expect(out['@type']).toBe('WebSite');
    expect(out.url).toBe('https://rtylr.com');
  });

  it('includes the site name and description', () => {
    const out = webSiteSchema();
    expect(out.name).toBe('rtylr');
    expect(typeof out.description).toBe('string');
  });
});

describe('webPageSchema', () => {
  it('emits a @graph with WebPage and BreadcrumbList', () => {
    const out = webPageSchema({ name: 'Blog', path: '/blog/', description: 'Blog posts.' });
    const graph = out['@graph'] as Array<Record<string, unknown>>;
    expect(graph).toHaveLength(2);
    expect(graph[0]['@type']).toBe('WebPage');
    expect(graph[1]['@type']).toBe('BreadcrumbList');
  });

  it('builds the correct WebPage URL', () => {
    const out = webPageSchema({ name: 'Support', path: '/support/', description: 'Help.' });
    const graph = out['@graph'] as Array<{ url?: string }>;
    expect(graph[0].url).toBe('https://rtylr.com/support/');
  });

  it('includes the page name and description', () => {
    const out = webPageSchema({ name: 'Trust', path: '/trust/', description: 'Security.' });
    const graph = out['@graph'] as Array<{ name?: string; description?: string }>;
    expect(graph[0].name).toBe('Trust');
    expect(graph[0].description).toBe('Security.');
  });
});
