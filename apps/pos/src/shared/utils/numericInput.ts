// Pure helpers for parsing and validating numeric user inputs.

/** Parse a string as a non-negative number; returns null on failure. */
export function parseNonNegativeNumber(input: string): number | null {
  const trimmed = input.trim();
  if (!trimmed) return null;
  const num = Number(trimmed);
  if (!Number.isFinite(num) || num < 0) return null;
  return num;
}

/** Parse a string as a positive number; returns null on failure. */
export function parsePositiveNumber(input: string): number | null {
  const num = parseNonNegativeNumber(input);
  if (num === null || num <= 0) return null;
  return num;
}

/** Parse a string as a positive integer; returns null on failure. */
export function parsePositiveInteger(input: string): number | null {
  const trimmed = input.trim();
  if (!trimmed) return null;
  const num = Number(trimmed);
  if (!Number.isFinite(num) || num <= 0) return null;
  return Math.floor(num);
}

/** Clamp a numeric input string to a [min, max] range.
 *  Returns the clamped value, or null if the input is not a valid number. */
export function clampInput(input: string, min: number, max: number): number | null {
  const num = Number(input.trim());
  if (!Number.isFinite(num)) return null;
  return Math.max(min, Math.min(max, num));
}

/** Return the number of decimal places in a numeric string. */
export function decimalPlaces(input: string): number {
  const dotIndex = input.indexOf('.');
  if (dotIndex === -1) return 0;
  return input.length - dotIndex - 1;
}

/** Format a number for display in a numeric keypad (avoid trailing zeros). */
export function keypadDisplay(value: number): string {
  if (!Number.isFinite(value)) return '0';
  if (Number.isInteger(value)) return String(value);
  return String(Math.round(value * 100) / 100);
}
