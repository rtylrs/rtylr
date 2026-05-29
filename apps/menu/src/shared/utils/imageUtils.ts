/**
 * Image and media helpers for the Menu app.
 */

/** Return true if the URL looks like a valid image URL. */
export function isImageUrl(url: string): boolean {
  if (!url) return false;
  return /\.(jpg|jpeg|png|gif|webp|avif|svg)(\?.*)?$/i.test(url);
}

/** Return a placeholder URL when no image is available. */
export function placeholderImage(width = 400, height = 300): string {
  return `https://placehold.co/${width}x${height}`;
}

/** Return an alt text string for a menu item image. */
export function menuItemAlt(itemName: string): string {
  return `Photo of ${itemName}`;
}

/** Build a srcset string from a base URL and width descriptors. */
export function buildSrcSet(baseUrl: string, widths: number[]): string {
  return widths.map((w) => `${baseUrl}?w=${w} ${w}w`).join(', ');
}

/** Return a CSS background-image string from a URL. */
export function cssBackgroundImage(url: string): string {
  return `url("${url}")`;
}

/** Clamp an image dimension to a min/max range. */
export function clampDimension(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

/** Convert a pixel dimension to a rem value (16px base). */
export function pxToRem(px: number, base = 16): string {
  return `${px / base}rem`;
}

/** Return an aspect-ratio CSS value from width and height. */
export function aspectRatioCss(width: number, height: number): string {
  return `${width} / ${height}`;
}
