// Color helpers for chart and UI theming in the insights feature.

const CHART_PALETTE = [
  '#6366f1',
  '#22d3ee',
  '#f59e0b',
  '#10b981',
  '#f43f5e',
  '#8b5cf6',
  '#fb923c',
  '#14b8a6',
];

export interface RgbColor {
  r: number;
  g: number;
  b: number;
}

/**
 * chartColor returns a palette color for the given index, wrapping around.
 */
export function chartColor(index: number): string {
  return CHART_PALETTE[index % CHART_PALETTE.length];
}

/**
 * assignColors maps each key to a chart palette color.
 */
export function assignColors(keys: string[]): Record<string, string> {
  const result: Record<string, string> = {};
  keys.forEach((key, i) => {
    result[key] = chartColor(i);
  });
  return result;
}

/**
 * hexToRgb converts a hex color string to an RgbColor.
 * Returns null when the input is not a valid 6-digit hex color.
 */
export function hexToRgb(hex: string): RgbColor | null {
  const cleaned = hex.startsWith('#') ? hex.slice(1) : hex;
  if (!/^[0-9a-fA-F]{6}$/.test(cleaned)) return null;
  return {
    r: parseInt(cleaned.slice(0, 2), 16),
    g: parseInt(cleaned.slice(2, 4), 16),
    b: parseInt(cleaned.slice(4, 6), 16),
  };
}

/**
 * relativeLuminance computes the WCAG relative luminance for a hex color.
 * Returns -1 when the hex color is invalid.
 */
export function relativeLuminance(hex: string): number {
  const rgb = hexToRgb(hex);
  if (!rgb) return -1;
  const toLinear = (c: number): number => {
    const s = c / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * toLinear(rgb.r) + 0.7152 * toLinear(rgb.g) + 0.0722 * toLinear(rgb.b);
}

/**
 * contrastingTextColor returns white or black based on WCAG luminance.
 */
export function contrastingTextColor(bgHex: string): '#ffffff' | '#000000' {
  const lum = relativeLuminance(bgHex);
  return lum < 0.179 ? '#ffffff' : '#000000';
}
