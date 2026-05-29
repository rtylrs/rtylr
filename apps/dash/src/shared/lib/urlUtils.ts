/**
 * URL and routing helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Joins path segments, handling leading/trailing slashes correctly. */
export function joinPath(...segments: string[]): string {
  return segments
    .map((s) => s.replace(/^\/+|\/+$/g, ''))
    .filter(Boolean)
    .join('/');
}

/** Returns the last segment of a URL path. */
export function lastPathSegment(path: string): string {
  const parts = path.replace(/\/+$/, '').split('/');
  return parts[parts.length - 1] ?? '';
}

/** Strips query string and hash from a URL. */
export function stripQueryAndHash(url: string): string {
  return url.split('?')[0].split('#')[0];
}

/** Returns the query string from a URL (without the leading ?). */
export function extractQueryString(url: string): string {
  const idx = url.indexOf('?');
  return idx === -1 ? '' : url.slice(idx + 1).split('#')[0];
}

/** Returns true if the given path matches or starts with the prefix. */
export function isActivePath(currentPath: string, prefix: string): boolean {
  if (prefix === '/') return currentPath === '/';
  return currentPath === prefix || currentPath.startsWith(prefix + '/');
}

/**
 * Builds a URL with additional query parameters merged into the existing ones.
 * Existing parameters are preserved and new ones are added or overwritten.
 */
export function withQueryParams(
  baseUrl: string,
  params: Record<string, string | number | boolean>
): string {
  const [path, existingQuery] = baseUrl.split('?');
  const search = new URLSearchParams(existingQuery ?? '');
  for (const [key, value] of Object.entries(params)) {
    search.set(key, String(value));
  }
  const qs = search.toString();
  return qs ? `${path}?${qs}` : path;
}

/** Returns true if the URL is absolute (starts with http:// or https://). */
export function isAbsoluteUrl(url: string): boolean {
  return /^https?:\/\//i.test(url);
}

/** Returns the origin (protocol + host) from a URL string, or null on failure. */
export function getOrigin(url: string): string | null {
  try {
    return new URL(url).origin;
  } catch {
    return null;
  }
}
