import { beforeEach, describe, expect, it } from 'vitest';

import { applyBrand, resolveIndex, resolveJob } from './resolver';
import type { Customization } from '../shared/lib/types';

function baseCustomization(overrides: Partial<Customization> = {}): Customization {
  return {
    id: 1,
    ownerId: 1,
    slug: 'acme',
    brandName: 'Acme',
    tagline: '',
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

describe('resolveIndex', () => {
  it('returns the correct component for each known template id', () => {
    expect(resolveIndex('minimal').name || resolveIndex('minimal').displayName).toBeDefined();
    // Not asserting on component identity (hot-reload would break
    // that); asserting that something non-null and callable is
    // returned per id.
    expect(typeof resolveIndex('minimal')).toBe('function');
    expect(typeof resolveIndex('editorial')).toBe('function');
    expect(typeof resolveIndex('bold')).toBe('function');
  });

  it('falls back to Minimal for an unknown template id', () => {
    // The resolver's fallback is Minimal rather than throwing, so a
    // stale DB value can't take down a business's public careers page.
    const unknown = resolveIndex('nonexistent' as never);
    const minimal = resolveIndex('minimal');
    expect(unknown).toBe(minimal);
  });
});

describe('resolveJob', () => {
  it('returns a function component for each known id', () => {
    expect(typeof resolveJob('minimal')).toBe('function');
    expect(typeof resolveJob('editorial')).toBe('function');
    expect(typeof resolveJob('bold')).toBe('function');
  });

  it('falls back to Minimal for an unknown id', () => {
    expect(resolveJob('nonexistent' as never)).toBe(resolveJob('minimal'));
  });
});

describe('applyBrand', () => {
  beforeEach(() => {
    // Strip any previously-set CSS vars so per-test state is clean.
    const root = document.documentElement;
    const propsToReset = [
      '--recruit-accent',
      '--recruit-accent-2',
      '--recruit-bg',
      '--recruit-fg',
      '--recruit-muted',
      '--recruit-surface',
      '--recruit-border',
      '--recruit-font-body',
      '--recruit-font-display',
    ];
    propsToReset.forEach((p) => root.style.removeProperty(p));
    // Strip any favicon link so the favicon-swap test has no stale
    // state.
    document.head.querySelectorAll('link[rel="icon"]').forEach((l) => l.remove());
    // Strip font link tags left over from prior tests.
    document.head.querySelectorAll('link[id^="recruit-font-"]').forEach((l) => l.remove());
  });

  it('is a no-op when called with undefined', () => {
    applyBrand(undefined);
    expect(document.documentElement.style.getPropertyValue('--recruit-accent')).toBe('');
  });

  it('sets accent to the primaryColor when provided', () => {
    applyBrand(baseCustomization({ primaryColor: '#ff00aa' }));
    expect(document.documentElement.style.getPropertyValue('--recruit-accent')).toBe('#ff00aa');
  });

  it('falls back to the template default when primaryColor is empty', () => {
    // The safe-default behaviour. Minimal's default accent is #111111
    // per the resolver; regression-check we don't ship blank CSS vars.
    applyBrand(baseCustomization({ primaryColor: '', template: 'minimal' }));
    expect(document.documentElement.style.getPropertyValue('--recruit-accent')).toBe('#111111');
  });

  it('uses accentColor for the secondary slot when provided', () => {
    applyBrand(
      baseCustomization({ primaryColor: '#000', accentColor: '#fff', template: 'minimal' })
    );
    expect(document.documentElement.style.getPropertyValue('--recruit-accent-2')).toBe('#fff');
  });

  it('falls accent-2 back to primaryColor when accentColor is empty', () => {
    applyBrand(baseCustomization({ primaryColor: '#abc', accentColor: '' }));
    expect(document.documentElement.style.getPropertyValue('--recruit-accent-2')).toBe('#abc');
  });

  it('picks the editorial template defaults for an editorial customization', () => {
    // Ensures the template-scoped fallback dictionary is wired up -
    // editorial bg is #ffffff not the neutral minimal #fafaf9.
    applyBrand(baseCustomization({ template: 'editorial' }));
    expect(document.documentElement.style.getPropertyValue('--recruit-bg')).toBe('#ffffff');
  });

  it('sets font CSS vars based on fontPairing', () => {
    applyBrand(baseCustomization({ fontPairing: 'playfair-inter' }));
    const display = document.documentElement.style.getPropertyValue('--recruit-font-display');
    const body = document.documentElement.style.getPropertyValue('--recruit-font-body');
    expect(display).toContain('Playfair');
    expect(body).toContain('Inter');
  });

  it('updates the favicon <link> when faviconUrl is set', () => {
    applyBrand(baseCustomization({ faviconUrl: 'https://cdn.test/f.ico' }));
    const link = document.querySelector('link[rel="icon"]') as HTMLLinkElement;
    expect(link?.href).toBe('https://cdn.test/f.ico');
  });

  it('creates a favicon <link> if none existed', () => {
    expect(document.querySelector('link[rel="icon"]')).toBeNull();
    applyBrand(baseCustomization({ faviconUrl: 'https://cdn.test/f.ico' }));
    expect(document.querySelector('link[rel="icon"]')).not.toBeNull();
  });

  it('does not touch the favicon when faviconUrl is empty', () => {
    applyBrand(baseCustomization({ faviconUrl: '' }));
    expect(document.querySelector('link[rel="icon"]')).toBeNull();
  });
});
