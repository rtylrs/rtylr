// Pure helpers for URL construction and query string manipulation.

/** Append query parameters to a base URL. */
export function buildUrl(base: string, params: Record<string, string | number | boolean>): string {
  const url = new URL(base, 'http://localhost');
  for (const [key, value] of Object.entries(params)) {
    url.searchParams.set(key, String(value));
  }
  return url.pathname + (url.search ? url.search : '');
}

/** Parse a query string into a key-value record. */
export function parseQuery(queryString: string): Record<string, string> {
  const params = new URLSearchParams(queryString.startsWith('?') ? queryString.slice(1) : queryString);
  const result: Record<string, string> = {};
  params.forEach((value, key) => {
    result[key] = value;
  });
  return result;
}

/** Return a single query parameter value, or null if not present. */
export function getQueryParam(queryString: string, key: string): string | null {
  const params = parseQuery(queryString);
  return params[key] ?? null;
}

/** Serialize a record into a query string (without leading "?"). */
export function serializeQuery(params: Record<string, string | number | boolean>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    search.set(key, String(value));
  }
  return search.toString();
}

/** Return true if a string looks like a valid absolute URL. */
export function isAbsoluteUrl(input: string): boolean {
  try {
    const url = new URL(input);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}
