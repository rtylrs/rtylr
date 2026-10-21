import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { applySEO, emitJobPostingJSONLD, loadAnalytics } from './seo';
import type { Customization, RecruitCareer } from './types';

/**
 * SEO helpers are silent: wrong meta tags or a broken JSON-LD only
 * surface in Search Console weeks later. These tests lock the DOM
 * effects so a regression flags immediately.
 */

function baseCustomization(overrides: Partial<Customization> = {}): Customization {
  return {
    id: 1,
    ownerId: 1,
    slug: 'acme',
    brandName: 'Acme',
    tagline: 'We hire',
    aboutBlurb: '',
    logoUrl: '',
    primaryColor: '',
    accentColor: '',
    bgColor: '',
    textColor: '',
    mutedTextColor: '',
    fontPairing: 'default',
    heroImageUrl: '',
    faviconUrl: '',
    template: 'minimal',
    headerHtml: '',
    metaTitle: '',
    metaDescription: '',
    privacyUrl: '',
    termsUrl: '',
    gaMeasurementId: '',
    metaPixelId: '',
    requireConsent: false,
    notificationEmail: '',
    emailNotifications: true,
    createdAt: '',
    updatedAt: '',
    ...overrides,
  };
}

function baseCareer(overrides: Partial<RecruitCareer> = {}): RecruitCareer {
  return {
    id: 1,
    ownerId: 1,
    slug: 'role',
    title: 'Senior Engineer',
    department: '',
    location: '',
    employmentType: '',
    workMode: '',
    seniority: '',
    details: 'Build things.',
    requirements: '',
    salaryMin: 0,
    salaryMax: 0,
    salaryCurrency: '',
    isPublished: true,
    isActive: true,
    createdAt: '2026-04-20T00:00:00Z',
    updatedAt: '2026-04-20T00:00:00Z',
    ...overrides,
  };
}

// Clean every DOM mutation between tests - applySEO reuses meta tags,
// loadAnalytics injects <script> tags, emitJobPostingJSONLD pins a
// specific script id. Without reset the tests cross-contaminate.
beforeEach(() => {
  document.head.innerHTML = '';
  document.title = '';
});
afterEach(() => {
  document.head.innerHTML = '';
});

describe('applySEO', () => {
  it('sets the document title from metaTitle when present', () => {
    applySEO(baseCustomization({ metaTitle: 'Custom title' }), { pageKind: 'index' });
    expect(document.title).toBe('Custom title');
  });

  it('builds a career-specific title on the career page', () => {
    applySEO(baseCustomization({ brandName: 'Acme' }), {
      pageKind: 'career',
      careerTitle: 'Senior Engineer',
    });
    expect(document.title).toBe('Senior Engineer - Acme');
  });

  it('falls back to "Careers at {brand}" when no metaTitle', () => {
    applySEO(baseCustomization({ brandName: 'Acme' }), { pageKind: 'index' });
    expect(document.title).toBe('Careers at Acme');
  });

  it('upserts og:title / og:description / og:image meta tags', () => {
    applySEO(baseCustomization({ brandName: 'Acme', tagline: 'Hire us' }), {
      pageKind: 'index',
    });
    const ogTitle = document.querySelector('meta[property="og:title"]') as HTMLMetaElement;
    const ogDesc = document.querySelector('meta[property="og:description"]') as HTMLMetaElement;
    const ogImage = document.querySelector('meta[property="og:image"]') as HTMLMetaElement;
    expect(ogTitle?.content).toBe('Careers at Acme');
    expect(ogDesc?.content).toContain('Hire us');
    expect(ogImage?.content).toContain('/recruit/public/og');
    expect(ogImage?.content).toContain('slug=acme');
  });

  it('adds twitter card meta tags', () => {
    applySEO(baseCustomization(), { pageKind: 'index' });
    const card = document.querySelector('meta[name="twitter:card"]') as HTMLMetaElement;
    expect(card?.content).toBe('summary_large_image');
  });

  it('updates the favicon <link> when faviconUrl is set', () => {
    applySEO(baseCustomization({ faviconUrl: 'https://cdn.test/favicon.ico' }), {
      pageKind: 'index',
    });
    const favicon = document.querySelector('link[rel="icon"]') as HTMLLinkElement;
    expect(favicon?.href).toBe('https://cdn.test/favicon.ico');
  });
});

describe('loadAnalytics', () => {
  // Analytics is gated on consent, so these cases grant it up front. The
  // gating itself is pinned in seo.analytics.test.ts.
  beforeEach(() => {
    document.cookie = 'rtylr-consent=granted; Path=/';
  });
  afterEach(() => {
    document.cookie = 'rtylr-consent=; Path=/; Max-Age=0';
  });

  it('injects GA4 scripts only when gaMeasurementId is set', () => {
    loadAnalytics(baseCustomization({ gaMeasurementId: 'G-ABC123' }));
    expect(document.getElementById('recruit-ga-G-ABC123')).not.toBeNull();
  });

  it('skips GA4 injection when gaMeasurementId is empty', () => {
    loadAnalytics(baseCustomization({ gaMeasurementId: '' }));
    const scripts = document.querySelectorAll('script[id^="recruit-ga-"]');
    expect(scripts.length).toBe(0);
  });

  it('injects Meta Pixel only when metaPixelId is set', () => {
    loadAnalytics(baseCustomization({ metaPixelId: '999' }));
    expect(document.getElementById('recruit-pixel-999')).not.toBeNull();
  });

  it('is idempotent - calling twice adds one tag per platform', () => {
    loadAnalytics(baseCustomization({ gaMeasurementId: 'G-X', metaPixelId: '1' }));
    loadAnalytics(baseCustomization({ gaMeasurementId: 'G-X', metaPixelId: '1' }));
    expect(document.querySelectorAll('script[id="recruit-ga-G-X"]').length).toBe(1);
    expect(document.querySelectorAll('script[id="recruit-pixel-1"]').length).toBe(1);
  });
});

describe('emitJobPostingJSONLD', () => {
  it('writes a JobPosting JSON-LD script with core fields', () => {
    emitJobPostingJSONLD(
      baseCareer({ title: 'Frontend', details: 'React work' }),
      baseCustomization({ brandName: 'Acme' })
    );
    const script = document.getElementById('recruit-jobposting-jsonld') as HTMLScriptElement;
    expect(script).not.toBeNull();
    expect(script.type).toBe('application/ld+json');
    const payload = JSON.parse(script.text);
    expect(payload['@type']).toBe('JobPosting');
    expect(payload.title).toBe('Frontend');
    expect(payload.description).toBe('React work');
    expect(payload.hiringOrganization.name).toBe('Acme');
  });

  it('omits baseSalary when salaryMin is 0', () => {
    emitJobPostingJSONLD(baseCareer({ salaryMin: 0 }), baseCustomization());
    const payload = JSON.parse(
      (document.getElementById('recruit-jobposting-jsonld') as HTMLScriptElement).text
    );
    expect(payload.baseSalary).toBeUndefined();
  });

  it('emits baseSalary range when only min is set (uses min as max)', () => {
    emitJobPostingJSONLD(
      baseCareer({ salaryMin: 80000, salaryMax: 0, salaryCurrency: 'USD' }),
      baseCustomization()
    );
    const payload = JSON.parse(
      (document.getElementById('recruit-jobposting-jsonld') as HTMLScriptElement).text
    );
    expect(payload.baseSalary.value.minValue).toBe(80000);
    expect(payload.baseSalary.value.maxValue).toBe(80000);
    expect(payload.baseSalary.currency).toBe('USD');
  });

  it('defaults currency to USD when unset', () => {
    emitJobPostingJSONLD(baseCareer({ salaryMin: 80000, salaryCurrency: '' }), baseCustomization());
    const payload = JSON.parse(
      (document.getElementById('recruit-jobposting-jsonld') as HTMLScriptElement).text
    );
    expect(payload.baseSalary.currency).toBe('USD');
  });

  it('replaces the previous JSON-LD block on re-emit', () => {
    // Simulates a route change from one role to another - we don't
    // want both JSON-LD blobs in the DOM at the same time.
    emitJobPostingJSONLD(baseCareer({ title: 'Old', slug: 'old' }), baseCustomization());
    emitJobPostingJSONLD(baseCareer({ title: 'New', slug: 'new' }), baseCustomization());
    const scripts = document.querySelectorAll('#recruit-jobposting-jsonld');
    expect(scripts.length).toBe(1);
    const payload = JSON.parse((scripts[0] as HTMLScriptElement).text);
    expect(payload.title).toBe('New');
  });

  it('emits employmentType only when the career has one', () => {
    emitJobPostingJSONLD(baseCareer({ employmentType: 'Full-time' }), baseCustomization());
    const payload = JSON.parse(
      (document.getElementById('recruit-jobposting-jsonld') as HTMLScriptElement).text
    );
    expect(payload.employmentType).toBe('Full-time');
  });
});
