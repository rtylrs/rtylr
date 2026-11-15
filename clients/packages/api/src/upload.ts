import { resolveCdnBaseUrl } from '@voxire/config';

export interface UploadedAsset {
  id: string;
  url: string;
  mime: string;
  size: number;
}

export interface UploadAssetOptions {
  cdnBaseUrl?: string;
  signal?: AbortSignal;
}

const DEFAULT_CDN_BASE_URL = 'https://api.voxire.com';

const getCdnBaseUrl = (override?: string): string => {
  const env = (import.meta as ImportMeta & { env?: Record<string, string | undefined> }).env;
  const base = override || env?.VITE_CDN_URL || resolveCdnBaseUrl() || DEFAULT_CDN_BASE_URL;
  return base.replace(/\/+$/, '');
};

const getToken = (): string | null => {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  let token: string | null = null;
  if (match) {
    try {
      token = decodeURIComponent(match[1]);
    } catch {
      // Malformed percent-encoding must not crash every upload. This mirrors
      // the guard its two siblings already have (client.ts getToken,
      // cookies.ts); this copy was the one that would throw URIError
      // synchronously and opaquely on a corrupted cookie (rtylr#952).
      token = match[1];
    }
  }
  if (!token || token.split('.').length !== 3) return null;
  return token;
};

const uploadErrorMessage = async (response: Response): Promise<string> => {
  const fallback = `Upload failed (${response.status})`;
  try {
    const body = await response.text();
    return body.trim() || fallback;
  } catch {
    return fallback;
  }
};

export const uploadAsset = async (
  file: File,
  options: UploadAssetOptions = {}
): Promise<UploadedAsset> => {
  const token = getToken();
  if (!token) {
    throw new Error('You need to sign in before uploading files.');
  }

  const body = new FormData();
  body.append('file', file);

  const response = await fetch(`${getCdnBaseUrl(options.cdnBaseUrl)}/upload`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body,
    credentials: 'include',
    signal: options.signal,
  });

  if (!response.ok) {
    throw new Error(await uploadErrorMessage(response));
  }

  return (await response.json()) as UploadedAsset;
};

export interface SignedAssetUrl {
  url: string;
  /** Unix seconds after which the link stops working. */
  expiresAt: number;
}

export interface SignAssetUrlOptions {
  /** Link lifetime in seconds. The CDN defaults to an hour and caps at 30 days. */
  expiresIn?: number;
  /**
   * Fixed expiry in Unix seconds, used instead of `expiresIn`. Signing every
   * view to the same expiry yields the same link, which the browser can cache.
   */
  expiresAt?: number;
  /** Image options covered by the signature, such as `{ w: '160' }`. */
  params?: Record<string, string>;
  cdnBaseUrl?: string;
  signal?: AbortSignal;
}

/**
 * Returns a time-limited link to one of the caller's own CDN files.
 *
 * Uploads are private by default, and the CDN serves a private file only to a
 * request that carries the user's token in an Authorization header, which an
 * `<img>` tag or a new tab cannot send. A signed link carries its own
 * credential, so it can be used directly as an image source or opened in a tab.
 */
export const signAssetUrl = async (
  url: string,
  options: SignAssetUrlOptions = {}
): Promise<SignedAssetUrl> => {
  const token = getToken();
  if (!token) {
    throw new Error('You need to sign in before opening files.');
  }

  const response = await fetch(`${getCdnBaseUrl(options.cdnBaseUrl)}/sign`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      url,
      ...(options.expiresAt
        ? { expires_at: options.expiresAt }
        : { expires_in: options.expiresIn ?? 3600 }),
      ...(options.params ? { params: options.params } : {}),
    }),
    credentials: 'include',
    signal: options.signal,
  });

  if (!response.ok) {
    throw new Error(await uploadErrorMessage(response));
  }

  const body = (await response.json()) as { url: string; expires_at: number };
  return { url: body.url, expiresAt: body.expires_at };
};
