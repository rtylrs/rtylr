// A POS till runs whatever keyboard layout the OS is set to, so an
// Arabic-locale cashier types U+0660-U+0669 where the app expects 0-9. Those
// characters are `\D` to an ASCII-only sanitiser, so stripping them discards
// the keystroke outright and the field never fills. Normalising instead keeps
// the value the app works with canonical ASCII.

// Code point of each block's zero; the next nine code points are its 1-9.
const DIGIT_BLOCK_STARTS = [
  0x0660, // Arabic-Indic
  0x06f0, // Eastern Arabic-Indic (Persian, Urdu)
  0x0966, // Devanagari
  0x09e6, // Bengali
  0x0a66, // Gurmukhi
  0x0ae6, // Gujarati
  0x0be6, // Tamil
  0x0c66, // Telugu
  0x0ce6, // Kannada
  0x0e50, // Thai
  0x0ed0, // Lao
  0xff10, // Fullwidth
];

const ASCII_DIGIT_MAP = new Map<string, string>();
for (const blockStart of DIGIT_BLOCK_STARTS) {
  for (let digit = 0; digit <= 9; digit += 1) {
    ASCII_DIGIT_MAP.set(String.fromCharCode(blockStart + digit), String(digit));
  }
}

/**
 * Rewrites non-ASCII decimal digits to ASCII and leaves every other character
 * untouched, so a genuinely non-numeric character is never translated.
 */
export function normalizeDigits(value: string): string {
  let normalized = '';
  // Iterates by code point, so astral characters survive intact.
  for (const character of value) {
    normalized += ASCII_DIGIT_MAP.get(character) ?? character;
  }
  return normalized;
}

/**
 * Normalises digits, then applies the existing policy of dropping anything
 * that is not a digit. `maxLength` caps the result, e.g. a 4-digit PIN.
 */
export function digitsOnly(value: string, maxLength?: number): string {
  const digits = normalizeDigits(value).replace(/\D/g, '');
  return maxLength === undefined ? digits : digits.slice(0, maxLength);
}
