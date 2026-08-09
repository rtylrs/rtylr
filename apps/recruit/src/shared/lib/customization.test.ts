import { describe, it, expect } from 'vitest';
import {
  primaryColor,
  accentColor,
  hasLogo,
  hasHeroImage,
  displayBrandName,
  careersMetaTitle,
  careersMetaDescription,
  isValidTemplate,
  effectiveTemplate,
  hasCustomFontPairing,
  buildCssVars,
  hasLegalLinks,
  DEFAULT_PRIMARY_COLOR,
  DEFAULT_ACCENT_COLOR,
} from './customization';
import type { Customization } from './types';

function makeCustomization(overrides: Partial<Customization> = {}): Customization {
  return {
    id: 1,
    ownerId: 1,
    slug: 'acme',
    brandName: 'Acme',
    tagline: 'We build things',
    aboutBlurb: '',
    logoUrl: '',
    primaryColor: '#111',
    accentColor: '#555',
    bgColor: '#fff',
    textColor: '#000',
    mutedTextColor: '#888',
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
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('primaryColor', () => {
  it('returns the customization primary color when set', () => {
    expect(primaryColor(makeCustomization({ primaryColor: '#ff0000' }))).toBe('#ff0000');
  });

  it('returns the default when primaryColor is empty', () => {
    expect(primaryColor(makeCustomization({ primaryColor: '' }))).toBe(DEFAULT_PRIMARY_COLOR);
  });
});

describe('accentColor', () => {
  it('returns the accent color when set', () => {
    expect(accentColor(makeCustomization({ accentColor: '#00ff00' }))).toBe('#00ff00');
  });

  it('returns the default when empty', () => {
    expect(accentColor(makeCustomization({ accentColor: '' }))).toBe(DEFAULT_ACCENT_COLOR);
  });
});

describe('hasLogo', () => {
  it('returns true when logoUrl is set', () => {
    expect(hasLogo(makeCustomization({ logoUrl: 'https://example.com/logo.png' }))).toBe(true);
  });

  it('returns false when logoUrl is empty', () => {
    expect(hasLogo(makeCustomization({ logoUrl: '' }))).toBe(false);
  });
});

describe('hasHeroImage', () => {
  it('returns true when heroImageUrl is set', () => {
    expect(hasHeroImage(makeCustomization({ heroImageUrl: 'https://example.com/hero.jpg' }))).toBe(
      true
    );
  });

  it('returns false when heroImageUrl is empty', () => {
    expect(hasHeroImage(makeCustomization({ heroImageUrl: '' }))).toBe(false);
  });
});

describe('displayBrandName', () => {
  it('returns brandName when set', () => {
    expect(displayBrandName(makeCustomization({ brandName: 'Acme Corp' }))).toBe('Acme Corp');
  });

  it('falls back to slug when brandName is empty', () => {
    expect(displayBrandName(makeCustomization({ brandName: '', slug: 'acme-corp' }))).toBe(
      'acme-corp'
    );
  });
});

describe('careersMetaTitle', () => {
  it('returns metaTitle when set', () => {
    expect(careersMetaTitle(makeCustomization({ metaTitle: 'Join Acme' }))).toBe('Join Acme');
  });

  it('constructs a title from brandName when metaTitle is empty', () => {
    expect(careersMetaTitle(makeCustomization({ metaTitle: '', brandName: 'Acme' }))).toBe(
      'Acme Careers'
    );
  });
});

describe('careersMetaDescription', () => {
  it('returns metaDescription when set', () => {
    expect(careersMetaDescription(makeCustomization({ metaDescription: 'Join our team' }))).toBe(
      'Join our team'
    );
  });

  it('falls back to tagline when metaDescription is empty', () => {
    expect(
      careersMetaDescription(makeCustomization({ metaDescription: '', tagline: 'We build things' }))
    ).toBe('We build things');
  });

  it('returns empty string when both are empty', () => {
    expect(careersMetaDescription(makeCustomization({ metaDescription: '', tagline: '' }))).toBe(
      ''
    );
  });
});

describe('isValidTemplate', () => {
  it('returns true for known templates', () => {
    expect(isValidTemplate('minimal')).toBe(true);
    expect(isValidTemplate('editorial')).toBe(true);
    expect(isValidTemplate('bold')).toBe(true);
  });

  it('returns false for unknown templates', () => {
    expect(isValidTemplate('dark')).toBe(false);
    expect(isValidTemplate('')).toBe(false);
  });
});

describe('effectiveTemplate', () => {
  it('returns the template when it is valid', () => {
    expect(effectiveTemplate(makeCustomization({ template: 'bold' }))).toBe('bold');
  });

  it('falls back to minimal for unknown templates', () => {
    expect(effectiveTemplate(makeCustomization({ template: 'futuristic' }))).toBe('minimal');
  });
});

describe('hasCustomFontPairing', () => {
  it('returns false for empty string', () => {
    expect(hasCustomFontPairing(makeCustomization({ fontPairing: '' }))).toBe(false);
  });

  it('returns false for default pairing', () => {
    expect(hasCustomFontPairing(makeCustomization({ fontPairing: 'default' }))).toBe(false);
  });

  it('returns true for a custom pairing', () => {
    expect(hasCustomFontPairing(makeCustomization({ fontPairing: 'jakarta-fraunces' }))).toBe(true);
  });
});

describe('buildCssVars', () => {
  it('includes the primary color variable', () => {
    const css = buildCssVars(makeCustomization({ primaryColor: '#abc' }));
    expect(css).toContain('--brand-primary: #abc');
  });

  it('includes all 5 CSS variables', () => {
    const css = buildCssVars(makeCustomization());
    expect(css.split(';').filter(Boolean).length).toBe(5);
  });
});

describe('hasLegalLinks', () => {
  it('returns true when both privacyUrl and termsUrl are set', () => {
    expect(
      hasLegalLinks(makeCustomization({ privacyUrl: 'https://a.com', termsUrl: 'https://b.com' }))
    ).toBe(true);
  });

  it('returns false when privacyUrl is missing', () => {
    expect(hasLegalLinks(makeCustomization({ privacyUrl: '', termsUrl: 'https://b.com' }))).toBe(
      false
    );
  });

  it('returns false when termsUrl is missing', () => {
    expect(hasLegalLinks(makeCustomization({ privacyUrl: 'https://a.com', termsUrl: '' }))).toBe(
      false
    );
  });
});
