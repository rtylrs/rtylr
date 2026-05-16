/** Normalize a URL by stripping a single trailing slash from the path. */
export function trimTrailingSlash(url: string): string {
  if (url.length > 1 && url.endsWith('/')) return url.slice(0, -1);
  return url;
}

/** Combine a base URL and path segment, deduplicating slashes between them. */
export function joinUrl(base: string, path: string): string {
  const b = base.endsWith('/') ? base.slice(0, -1) : base;
  const p = path.startsWith('/') ? path : '/' + path;
  return b + p;
}
