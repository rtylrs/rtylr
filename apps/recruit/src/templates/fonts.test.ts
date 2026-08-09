import { beforeEach, describe, expect, it } from 'vitest';

import { FONT_PAIRINGS, loadFontPairing, resolvePairing } from './fonts';

/**
 * Font resolver tests - these guard the resolver against silently
 * falling back to the wrong pairing (one of the longest-to-debug
 * classes of UI bugs: the whole page renders in Plus Jakarta because
 * a typo in a font key).
 */

describe('resolvePairing', () => {
  it('returns default pairing for undefined input', () => {
    expect(resolvePairing(undefined).id).toBe('default');
  });

  it('returns default pairing for empty string', () => {
    expect(resolvePairing('').id).toBe('default');
  });

  it('returns default pairing for an unknown id', () => {
    // A stale database value shouldn't render a broken font stack -
    // fall through to the safe default.
    expect(resolvePairing('jakarta-fruncis').id).toBe('default');
  });

  it('returns the requested pairing when it exists', () => {
    expect(resolvePairing('editorial' as never).id).not.toBe('editorial');
    expect(resolvePairing('jakarta-fraunces').id).toBe('jakarta-fraunces');
    expect(resolvePairing('playfair-inter').id).toBe('playfair-inter');
  });

  it('every declared pairing is self-consistent (id matches key)', () => {
    // Catches a copy-paste bug where a new pairing's `id` gets set
    // to the name of the pairing it was copied from.
    for (const [key, pair] of Object.entries(FONT_PAIRINGS)) {
      expect(pair.id).toBe(key);
    }
  });

  it('every pairing declares heading, body, and googleFonts', () => {
    for (const pair of Object.values(FONT_PAIRINGS)) {
      expect(pair.heading).toMatch(/^["'].+["']$/);
      expect(pair.body).toMatch(/^["'].+["']$/);
      expect(pair.googleFonts).toMatch(/^family=/);
    }
  });
});

describe('loadFontPairing', () => {
  beforeEach(() => {
    // Reset between tests - the function injects <link> tags into
    // document.head and we need each test to start from a clean slate.
    document.head.querySelectorAll('link[id^="recruit-font-"]').forEach((el) => el.remove());
  });

  it('injects a Google Fonts link on first call', () => {
    loadFontPairing('jakarta-fraunces');
    const link = document.getElementById('recruit-font-jakarta-fraunces') as HTMLLinkElement | null;
    expect(link).not.toBeNull();
    expect(link?.href).toContain('fonts.googleapis.com');
    expect(link?.href).toContain('Fraunces');
  });

  it('is idempotent - calling twice only injects one tag', () => {
    loadFontPairing('jakarta-fraunces');
    loadFontPairing('jakarta-fraunces');
    const links = document.head.querySelectorAll('link[id="recruit-font-jakarta-fraunces"]');
    expect(links.length).toBe(1);
  });

  it('falls back to default for unknown ids without injecting the wrong link', () => {
    loadFontPairing('nonexistent');
    // Should have injected the default font's link, not the missing one.
    expect(document.getElementById('recruit-font-default')).not.toBeNull();
    expect(document.getElementById('recruit-font-nonexistent')).toBeNull();
  });

  it('returns the resolved pairing regardless of input', () => {
    const result = loadFontPairing('playfair-inter');
    expect(result.id).toBe('playfair-inter');
    expect(result.heading).toContain('Playfair');
  });
});
