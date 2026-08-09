/**
 * Attempts to parse a URL string. Returns null if the value is not a valid URL.
 */
export function safeParseUrl(value: string): URL | null {
  try {
    return new URL(value);
  } catch {
    return null;
  }
}

/**
 * Extracts a named search parameter from a full URL string or a bare query string.
 * Returns null when the parameter is absent.
 */
export function getSearchParam(urlOrSearch: string, name: string): string | null {
  const parsed = safeParseUrl(urlOrSearch);
  if (parsed) return parsed.searchParams.get(name);
  try {
    const params = new URLSearchParams(urlOrSearch);
    return params.get(name);
  } catch {
    return null;
  }
}

/**
 * Appends or overwrites a search parameter on a URL string.
 * Returns the original string unchanged when the input is not a valid URL.
 */
export function appendSearchParam(url: string, name: string, value: string): string {
  const parsed = safeParseUrl(url);
  if (!parsed) return url;
  parsed.searchParams.set(name, value);
  return parsed.toString();
}

/**
 * Returns true when the destination is safe to redirect to.
 * A destination is considered safe when it shares the current origin or matches
 * one of the explicitly provided allowed origins.
 */
export function isSafeRedirect(destination: string, allowedOrigins: string[] = []): boolean {
  if (!destination) return false;
  const parsed = safeParseUrl(destination);
  if (!parsed) return false;
  const currentOrigin = typeof window !== 'undefined' ? window.location.origin : '';
  const permitted = [currentOrigin, ...allowedOrigins];
  return permitted.some((origin) => origin && parsed.origin === origin);
}

/**
 * Removes a trailing slash from a path, unless the path is the root slash.
 */
export function stripTrailingSlash(path: string): string {
  if (path === '/') return path;
  return path.endsWith('/') ? path.slice(0, -1) : path;
}
