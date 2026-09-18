/**
 * Color utility helpers for the landing app.
 * Supports hex, RGB, HSL conversions and accessibility checks.
 */

/** Converts a hex color string to an RGB object. */
export function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const clean = hex.replace(/^#/, '');
  if (clean.length !== 6 && clean.length !== 3) return null;
  const full =
    clean.length === 3
      ? clean
          .split('')
          .map((c) => c + c)
          .join('')
      : clean;
  const num = parseInt(full, 16);
  if (isNaN(num)) return null;
  return {
    r: (num >> 16) & 0xff,
    g: (num >> 8) & 0xff,
    b: num & 0xff,
  };
}

/** Converts RGB values to a hex color string. */
export function rgbToHex(r: number, g: number, b: number): string {
  return (
    '#' +
    [r, g, b]
      .map((v) => {
        const clamped = Math.max(0, Math.min(255, Math.round(v)));
        return clamped.toString(16).padStart(2, '0');
      })
      .join('')
  );
}

/** Returns the relative luminance of an RGB color (0 to 1). */
export function relativeLuminance(r: number, g: number, b: number): number {
  const channel = (c: number) => {
    const s = c / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b);
}

/** Returns the WCAG contrast ratio between two hex colors. */
export function contrastRatio(hex1: string, hex2: string): number {
  const c1 = hexToRgb(hex1);
  const c2 = hexToRgb(hex2);
  if (!c1 || !c2) return 1;
  const l1 = relativeLuminance(c1.r, c1.g, c1.b);
  const l2 = relativeLuminance(c2.r, c2.g, c2.b);
  const lighter = Math.max(l1, l2);
  const darker = Math.min(l1, l2);
  return (lighter + 0.05) / (darker + 0.05);
}

/** Returns true when the contrast ratio meets WCAG AA standard (>= 4.5). */
export function isAccessibleAA(hex1: string, hex2: string): boolean {
  return contrastRatio(hex1, hex2) >= 4.5;
}

/** Returns true when the contrast ratio meets WCAG AAA standard (>= 7). */
export function isAccessibleAAA(hex1: string, hex2: string): boolean {
  return contrastRatio(hex1, hex2) >= 7;
}

/** Returns '#000000' or '#ffffff' depending on which contrasts better with the background. */
export function readableTextColor(backgroundHex: string): string {
  const rgb = hexToRgb(backgroundHex);
  if (!rgb) return '#000000';
  const lum = relativeLuminance(rgb.r, rgb.g, rgb.b);
  return lum > 0.179 ? '#000000' : '#ffffff';
}

/** Lightens a hex color by a percentage (0 to 100). */
export function lighten(hex: string, amount: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;
  const factor = amount / 100;
  return rgbToHex(
    rgb.r + (255 - rgb.r) * factor,
    rgb.g + (255 - rgb.g) * factor,
    rgb.b + (255 - rgb.b) * factor
  );
}

/** Darkens a hex color by a percentage (0 to 100). */
export function darken(hex: string, amount: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;
  const factor = 1 - amount / 100;
  return rgbToHex(rgb.r * factor, rgb.g * factor, rgb.b * factor);
}

/** Returns a hex color with the given opacity applied on top of white. */
export function withAlphaOnWhite(hex: string, alpha: number): string {
  const rgb = hexToRgb(hex);
  if (!rgb) return hex;
  const a = Math.max(0, Math.min(1, alpha));
  return rgbToHex(
    Math.round(rgb.r * a + 255 * (1 - a)),
    Math.round(rgb.g * a + 255 * (1 - a)),
    Math.round(rgb.b * a + 255 * (1 - a))
  );
}
