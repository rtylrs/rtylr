// Pure string manipulation helpers used across the POS UI.

/** Capitalize the first letter of a string. */
export function capitalize(text: string): string {
  if (!text) return '';
  return text.charAt(0).toUpperCase() + text.slice(1);
}

/** Convert a snake_case or kebab-case string to Title Case. */
export function toTitleCase(text: string): string {
  return text
    .replace(/[-_]/g, ' ')
    .split(' ')
    .map((word) => (word ? word.charAt(0).toUpperCase() + word.slice(1).toLowerCase() : ''))
    .join(' ')
    .trim();
}

/** Return initials from a full name (up to two letters). */
export function initials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '';
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
}

/** Mask all but the last N characters of a string with a given mask char. */
export function maskExceptLast(text: string, visibleCount = 4, maskChar = '*'): string {
  if (text.length <= visibleCount) return text;
  const visible = text.slice(-visibleCount);
  const masked = maskChar.repeat(text.length - visibleCount);
  return masked + visible;
}

/** Return true if a string is a valid email address. */
export function isEmail(text: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(text.trim());
}

/** Return true if a string contains only digits. */
export function isNumericString(text: string): boolean {
  return /^\d+$/.test(text.trim());
}

/** Remove non-numeric characters from a string (useful for phone numbers). */
export function digitsOnly(text: string): string {
  return text.replace(/\D/g, '');
}

/** Normalize whitespace: trim and collapse internal spaces to single spaces. */
export function normalizeWhitespace(text: string): string {
  return text.trim().replace(/\s+/g, ' ');
}
