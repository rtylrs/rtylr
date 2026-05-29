/**
 * String manipulation helpers for the landing app.
 * Covers truncation, slug generation, title casing, and template interpolation.
 */

/**
 * Truncates a string to a maximum length, appending an ellipsis if cut.
 * truncate('hello world', 8) -> 'hello...'
 */
export function truncate(text: string, maxLength: number, ellipsis = '...'): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength - ellipsis.length) + ellipsis;
}

/**
 * Converts a string to a URL-safe slug.
 * slugify('Hello, World!') -> 'hello-world'
 */
export function slugify(text: string): string {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * Converts a slug back to a title-cased string.
 * deslugify('hello-world') -> 'Hello World'
 */
export function deslugify(slug: string): string {
  return slug
    .split('-')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

/**
 * Title-cases a string: each word's first letter is uppercased, the rest lowercased.
 * toTitleCase('the quick brown fox') -> 'The Quick Brown Fox'
 */
export function toTitleCase(text: string): string {
  return text.replace(
    /\w\S*/g,
    (word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase()
  );
}

/**
 * Capitalizes only the first character of a string.
 * capitalize('hello world') -> 'Hello world'
 */
export function capitalize(text: string): string {
  if (!text) return text;
  return text.charAt(0).toUpperCase() + text.slice(1);
}

/**
 * Interpolates named placeholders in a template string.
 * interpolate('Hello, {name}!', { name: 'World' }) -> 'Hello, World!'
 */
export function interpolate(template: string, vars: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (match, key) => {
    return Object.prototype.hasOwnProperty.call(vars, key) ? String(vars[key]) : match;
  });
}

/**
 * Returns a string with all runs of whitespace collapsed to a single space and
 * leading/trailing whitespace removed.
 */
export function collapseWhitespace(text: string): string {
  return text.replace(/\s+/g, ' ').trim();
}

/**
 * Strips HTML tags from a string.
 * stripHtml('<p>Hello <b>world</b></p>') -> 'Hello world'
 */
export function stripHtml(html: string): string {
  return html.replace(/<[^>]*>/g, '');
}

/**
 * Counts the number of words in a string.
 */
export function wordCount(text: string): number {
  return text.trim() === '' ? 0 : text.trim().split(/\s+/).length;
}

/**
 * Estimates reading time in minutes for a given text.
 * Uses an average reading speed of 200 words per minute.
 */
export function readingTimeMinutes(text: string, wpm = 200): number {
  return Math.ceil(wordCount(text) / wpm);
}
