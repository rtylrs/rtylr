import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { loadAnalytics } from './seo';
import type { Customization } from './types';

/**
 * The analytics identifiers are business-supplied free text and these pages
 * are anonymous, so the loader has two jobs: never let an identifier become
 * script source, and never fire before the visitor consents.
 */

function customization(overrides: Partial<Customization> = {}): Customization {
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

function setConsent(value: 'granted' | 'denied' | ''): void {
  document.cookie = 'rtylr-consent=; Path=/; Max-Age=0';
  if (value) document.cookie = `rtylr-consent=${value}; Path=/`;
}

function announceConsent(): void {
  window.dispatchEvent(new CustomEvent('rtylr-consent-change'));
}

function analyticsScripts(): HTMLScriptElement[] {
  return Array.from(
    document.head.querySelectorAll('script[id^="recruit-ga-"], script[id^="recruit-pixel-"]')
  );
}

function scriptById(id: string): HTMLScriptElement | null {
  return document.getElementById(id) as HTMLScriptElement | null;
}

beforeEach(() => {
  document.head.innerHTML = '';
  setConsent('granted');
});
afterEach(() => {
  document.head.innerHTML = '';
  setConsent('');
});

describe('loadAnalytics - valid identifiers', () => {
  it('emits the GA4 loader and init for a well-formed measurement id', () => {
    loadAnalytics(customization({ gaMeasurementId: 'G-ABC123XYZ9' }));

    const loader = scriptById('recruit-ga-G-ABC123XYZ9');
    expect(loader).not.toBeNull();
    expect(loader?.src).toContain('googletagmanager.com/gtag/js?id=G-ABC123XYZ9');

    const init = scriptById('recruit-ga-G-ABC123XYZ9-init');
    expect(init?.getAttribute('data-recruit-ga')).toBe('G-ABC123XYZ9');
  });

  it('emits the Meta Pixel for a numeric pixel id', () => {
    loadAnalytics(customization({ metaPixelId: '123456789012345' }));

    const script = scriptById('recruit-pixel-123456789012345');
    expect(script).not.toBeNull();
    expect(script?.getAttribute('data-recruit-pixel')).toBe('123456789012345');
  });

  it('never concatenates the identifier into the inline script source', () => {
    loadAnalytics(customization({ gaMeasurementId: 'G-AAA111', metaPixelId: '111' }));
    const gaSource = scriptById('recruit-ga-G-AAA111-init')?.text ?? '';
    const pixelSource = scriptById('recruit-pixel-111')?.text ?? '';
    expect(gaSource).not.toContain('G-AAA111');
    expect(pixelSource).not.toContain('111');

    // Same invariant the other way round: two different valid ids must produce
    // byte-identical script bodies, because the id is not part of the source.
    loadAnalytics(customization({ gaMeasurementId: 'G-BBB222' }));
    expect(scriptById('recruit-ga-G-BBB222-init')?.text).toBe(gaSource);
  });
});

describe('loadAnalytics - hostile identifiers', () => {
  it('emits nothing for a script-breakout GA4 id', () => {
    loadAnalytics(customization({ gaMeasurementId: "G-ABC';alert(document.cookie);//" }));
    expect(analyticsScripts()).toHaveLength(0);
  });

  it('emits nothing for a script-breakout Meta Pixel id', () => {
    loadAnalytics(customization({ metaPixelId: "1');alert(document.cookie);fbq('x" }));
    expect(analyticsScripts()).toHaveLength(0);
  });

  it('emits nothing for ids that close the loader tag or smuggle markup', () => {
    const payloads = [
      'G-ABC<script>alert(1)</script>',
      'G-ABC" onload="alert(1)',
      'G-ABC\n</script><script>alert(1)</script>',
      '</script><script>alert(1)</script>',
      'ABC123',
      '',
    ];
    for (const gaMeasurementId of payloads) {
      document.head.innerHTML = '';
      loadAnalytics(customization({ gaMeasurementId }));
      expect(analyticsScripts()).toHaveLength(0);
    }
  });

  it('emits nothing for non-numeric Meta Pixel ids', () => {
    const payloads = ["123',alert(1);//", 'abc123', '123-456', 'G-123', ' '];
    for (const metaPixelId of payloads) {
      document.head.innerHTML = '';
      loadAnalytics(customization({ metaPixelId }));
      expect(analyticsScripts()).toHaveLength(0);
    }
  });
});

describe('loadAnalytics - consent', () => {
  it('emits nothing when consent is undecided', () => {
    setConsent('');
    loadAnalytics(customization({ gaMeasurementId: 'G-ABC123', metaPixelId: '999' }));
    expect(analyticsScripts()).toHaveLength(0);
  });

  it('emits nothing when consent is denied', () => {
    setConsent('denied');
    loadAnalytics(customization({ gaMeasurementId: 'G-ABC123', metaPixelId: '999' }));
    expect(analyticsScripts()).toHaveLength(0);
  });

  it('loads once the visitor consents after the page rendered', () => {
    setConsent('');
    loadAnalytics(customization({ gaMeasurementId: 'G-ABC123', metaPixelId: '999' }));
    expect(analyticsScripts()).toHaveLength(0);

    setConsent('granted');
    announceConsent();
    expect(scriptById('recruit-ga-G-ABC123')).not.toBeNull();
    expect(scriptById('recruit-pixel-999')).not.toBeNull();
  });

  it('drops the tags again when consent is withdrawn', () => {
    loadAnalytics(customization({ gaMeasurementId: 'G-ABC123', metaPixelId: '999' }));
    expect(analyticsScripts()).toHaveLength(3);

    setConsent('denied');
    announceConsent();
    expect(analyticsScripts()).toHaveLength(0);
  });

  it('does not re-add a tag that was already removed', () => {
    setConsent('');
    loadAnalytics(customization({ gaMeasurementId: 'G-ABC123' }));
    setConsent('granted');
    announceConsent();
    loadAnalytics(customization({ gaMeasurementId: 'G-ABC123' }));
    expect(document.querySelectorAll('script[id="recruit-ga-G-ABC123"]').length).toBe(1);
  });
});
