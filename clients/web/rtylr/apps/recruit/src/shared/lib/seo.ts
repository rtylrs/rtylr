import type { Customization, RecruitCareer } from './types';

/**
 * SEO / analytics helpers for the public careers pages. Everything is
 * DOM-side because the app is a SPA - we don't have SSR. That means:
 *   - <title> is set imperatively
 *   - meta tags are upserted (create-or-update)
 *   - GA4 and Meta Pixel scripts are injected once, guarded by marker
 *     ids so navigating between pages doesn't stack them
 *   - JobPosting structured data is injected per-role
 *
 * Missing IDs are a no-op; a business that never pastes a GA ID sees
 * zero tracking script on their page. That's the design - opt-in.
 */

function upsertMeta(attrKey: 'name' | 'property', attrValue: string, content: string): void {
  if (!content) return;
  const selector = `meta[${attrKey}="${attrValue}"]`;
  let el = document.querySelector<HTMLMetaElement>(selector);
  if (!el) {
    el = document.createElement('meta');
    el.setAttribute(attrKey, attrValue);
    document.head.appendChild(el);
  }
  el.setAttribute('content', content);
}

function upsertLink(rel: string, href: string): void {
  if (!href) return;
  let el = document.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`);
  if (!el) {
    el = document.createElement('link');
    el.setAttribute('rel', rel);
    document.head.appendChild(el);
  }
  el.setAttribute('href', href);
}

/**
 * applySEO sets title + meta for a public careers view. Call this on
 * page mount (alongside applyBrand); each field falls back to a
 * derivation of brand_name + tagline so unconfigured rows still get
 * a sensible social preview.
 */
export function applySEO(
  customization: Customization,
  context?: { careerTitle?: string; pageKind: 'index' | 'career' }
): void {
  if (typeof document === 'undefined') return;

  const brand = customization.brandName || 'Careers';
  const title =
    context?.pageKind === 'career' && context.careerTitle
      ? `${context.careerTitle} - ${brand}`
      : customization.metaTitle || `Careers at ${brand}`;
  document.title = title;

  const description =
    customization.metaDescription ||
    customization.tagline ||
    `Open roles at ${brand}. Apply online in a few minutes.`;

  upsertMeta('name', 'description', description);
  upsertMeta('property', 'og:title', title);
  upsertMeta('property', 'og:description', description);
  upsertMeta('property', 'og:type', 'website');

  // OG image - the server-rendered preview at /recruit/public/og. The
  // slug query is set on every call so LinkedIn's crawler gets a fresh
  // image after branding edits.
  const apiBase = (import.meta.env.VITE_API_URL as string | undefined) ?? 'https://api.voxire.com';
  const ogImage = `${apiBase.replace(/\/$/, '')}/recruit/public/og?slug=${encodeURIComponent(customization.slug)}`;
  upsertMeta('property', 'og:image', ogImage);
  upsertMeta('name', 'twitter:card', 'summary_large_image');
  upsertMeta('name', 'twitter:title', title);
  upsertMeta('name', 'twitter:description', description);
  upsertMeta('name', 'twitter:image', ogImage);

  if (customization.faviconUrl) upsertLink('icon', customization.faviconUrl);
}

/**
 * loadAnalytics injects GA4 and / or Meta Pixel loaders based on the
 * customization. Each is gated by a marker id so repeat calls across
 * route changes don't re-add the script tags.
 */
export function loadAnalytics(customization: Customization): void {
  if (typeof document === 'undefined') return;

  if (customization.gaMeasurementId) {
    const gaID = customization.gaMeasurementId.trim();
    const markerID = `recruit-ga-${gaID}`;
    if (!document.getElementById(markerID)) {
      const loader = document.createElement('script');
      loader.id = markerID;
      loader.async = true;
      loader.src = `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(gaID)}`;
      document.head.appendChild(loader);

      const init = document.createElement('script');
      init.text = `window.dataLayer = window.dataLayer || [];function gtag(){dataLayer.push(arguments);}gtag('js', new Date());gtag('config', '${gaID}');`;
      document.head.appendChild(init);
    }
  }

  if (customization.metaPixelId) {
    const pixelID = customization.metaPixelId.trim();
    const markerID = `recruit-pixel-${pixelID}`;
    if (!document.getElementById(markerID)) {
      const script = document.createElement('script');
      script.id = markerID;
      script.text = `!function(f,b,e,v,n,t,s){if(f.fbq)return;n=f.fbq=function(){n.callMethod?n.callMethod.apply(n,arguments):n.queue.push(arguments)};if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';n.queue=[];t=b.createElement(e);t.async=!0;t.src=v;s=b.getElementsByTagName(e)[0];s.parentNode.insertBefore(t,s)}(window,document,'script','https://connect.facebook.net/en_US/fbevents.js');fbq('init', '${pixelID}');fbq('track', 'PageView');`;
      document.head.appendChild(script);
    }
  }
}

/**
 * emitJobPostingJSONLD drops a <script type="application/ld+json"> tag
 * on the page describing the role in schema.org JobPosting shape.
 * Google Jobs ingests these and surfaces listings in search.
 * Idempotent - removes the previous block before inserting the new one.
 */
export function emitJobPostingJSONLD(career: RecruitCareer, customization: Customization): void {
  if (typeof document === 'undefined') return;

  const id = 'recruit-jobposting-jsonld';
  const existing = document.getElementById(id);
  if (existing) existing.remove();

  const apiBase = (import.meta.env.VITE_API_URL as string | undefined) ?? 'https://api.voxire.com';
  const url = `${typeof window !== 'undefined' ? window.location.origin : ''}/${customization.slug}/${career.slug}`;

  // Conservative JSON-LD - only emit fields we actually have values
  // for. Google tolerates missing recommended fields, but strings
  // filled with empty placeholders look bad in search snippets.
  const payload: Record<string, unknown> = {
    '@context': 'https://schema.org',
    '@type': 'JobPosting',
    title: career.title,
    description: career.details,
    datePosted: career.createdAt,
    hiringOrganization: {
      '@type': 'Organization',
      name: customization.brandName,
      sameAs: apiBase,
      logo: customization.logoUrl || undefined,
    },
    url,
  };

  if (career.location) {
    payload.jobLocation = {
      '@type': 'Place',
      address: { '@type': 'PostalAddress', addressLocality: career.location },
    };
  }
  if (career.employmentType) payload.employmentType = career.employmentType;
  if (career.salaryMin > 0) {
    payload.baseSalary = {
      '@type': 'MonetaryAmount',
      currency: career.salaryCurrency || 'USD',
      value: {
        '@type': 'QuantitativeValue',
        minValue: career.salaryMin,
        maxValue: career.salaryMax > 0 ? career.salaryMax : career.salaryMin,
        unitText: 'YEAR',
      },
    };
  }

  const script = document.createElement('script');
  script.id = id;
  script.type = 'application/ld+json';
  script.text = JSON.stringify(payload);
  document.head.appendChild(script);
}
