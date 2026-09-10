import type { FontPairingId } from '../shared/lib/types';

/**
 * FONT_PAIRINGS is the complete set of typography choices a business
 * can pick in the admin. We deliberately don't expose a freeform
 * Google Fonts picker - too many bad pairings, and the template
 * designs assume specific proportions. Adding a new pairing means
 * editing this file and refreshing in the admin picker.
 *
 * Each pairing declares:
 *   - heading: the display-face family name, used in CSS.
 *   - body:    the supporting face used for paragraph text.
 *   - fonts: the face names, shown to an admin beside a translated
 *            descriptor. Typeface names are proper nouns, so they are not
 *            translated and stay here rather than in a locale file.
 *   - googleFonts: the `family=...` query-string fragment that loads
 *                  both faces from fonts.googleapis.com. Combined with
 *                  the base URL in loadFontPairing().
 */
interface FontPairing {
  id: FontPairingId;
  fonts: string;
  heading: string;
  body: string;
  googleFonts: string;
}

export const FONT_PAIRINGS: Record<Exclude<FontPairingId, ''>, FontPairing> = {
  default: {
    id: 'default',
    fonts: 'Plus Jakarta Sans',
    heading: '"Plus Jakarta Sans"',
    body: '"Plus Jakarta Sans"',
    googleFonts: 'family=Plus+Jakarta+Sans:wght@400;500;600;700',
  },
  'jakarta-fraunces': {
    id: 'jakarta-fraunces',
    fonts: 'Fraunces + Plus Jakarta',
    heading: '"Fraunces"',
    body: '"Plus Jakarta Sans"',
    googleFonts:
      'family=Fraunces:opsz,wght@9..144,600..800&family=Plus+Jakarta+Sans:wght@400;500;600',
  },
  'space-grotesk-inter': {
    id: 'space-grotesk-inter',
    fonts: 'Space Grotesk + Inter',
    heading: '"Space Grotesk"',
    body: '"Inter"',
    googleFonts: 'family=Space+Grotesk:wght@500..700&family=Inter:wght@400;500',
  },
  'playfair-inter': {
    id: 'playfair-inter',
    fonts: 'Playfair + Inter',
    heading: '"Playfair Display"',
    body: '"Inter"',
    googleFonts: 'family=Playfair+Display:wght@600..900&family=Inter:wght@400;500',
  },
  'dmserif-dmsans': {
    id: 'dmserif-dmsans',
    fonts: 'DM Serif + DM Sans',
    heading: '"DM Serif Display"',
    body: '"DM Sans"',
    googleFonts: 'family=DM+Serif+Display&family=DM+Sans:wght@400;500;600',
  },
};

/**
 * resolvePairing returns the pairing to apply for a given customization
 * field. Unknown / empty values fall back to `default` so every page
 * renders with coherent typography regardless of what's in the DB.
 */
export function resolvePairing(id: string | undefined): FontPairing {
  if (id && id in FONT_PAIRINGS) {
    return FONT_PAIRINGS[id as keyof typeof FONT_PAIRINGS];
  }
  return FONT_PAIRINGS.default;
}

/**
 * loadFontPairing injects a <link rel="stylesheet"> tag pointing at
 * fonts.googleapis.com. Idempotent - if the same pairing is already
 * loaded, it's a no-op. Switching pairings leaves stale link tags in
 * the DOM which is fine: Google Fonts caches aggressively and the
 * stale tag won't affect rendering.
 */
export function loadFontPairing(id: string | undefined): FontPairing {
  const pair = resolvePairing(id);
  if (typeof document === 'undefined') return pair;

  const markerID = `recruit-font-${pair.id}`;
  if (document.getElementById(markerID)) return pair;

  const link = document.createElement('link');
  link.id = markerID;
  link.rel = 'stylesheet';
  link.href = `https://fonts.googleapis.com/css2?${pair.googleFonts}&display=swap`;
  document.head.appendChild(link);
  return pair;
}
