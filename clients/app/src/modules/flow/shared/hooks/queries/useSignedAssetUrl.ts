import { signAssetUrl, type SignedAssetUrl } from '@voxire/api';
import { getCdnBaseUrl } from '@voxire/config';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';

const HOUR_SECONDS = 60 * 60;
// A link is re-signed once it has less than this left, so an image still on
// screen never points at an expired link.
const RESIGN_MARGIN_SECONDS = 10 * 60;

const KNOWN_CDN_HOSTS = new Set([
  'cdn.voxire.com',
  'cdn.voxire.dev',
  'api.voxire.com',
  'api.voxire.dev',
]);

/**
 * Which copy of a file to load. A thumbnail or preview is a resized copy, so a
 * 64px thumbnail does not download a multi-megabyte screenshot; the original
 * is the untouched upload.
 */
export type AssetVariant = 'thumbnail' | 'preview' | 'original';

const VARIANT_WIDTH: Record<Exclude<AssetVariant, 'original'>, string> = {
  // Twice the 64px frame, for high-density screens.
  thumbnail: '160',
  preview: '1600',
};

const FILE_PATH = /^\/files\/([^/?#]+)$/;

/**
 * Whether a stored address names a file on our CDN. Those files are private by
 * default and need a signed link to load in an `<img>` or a new tab. Anything
 * else (an external link someone attached) is used as stored.
 */
export function isCdnFileUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    const cdnHost = new URL(getCdnBaseUrl()).host.toLowerCase();
    const host = parsed.host.toLowerCase();
    const onCdn = host === cdnHost || KNOWN_CDN_HOSTS.has(parsed.hostname.toLowerCase());
    const filePath =
      parsed.pathname.startsWith('/files/') ||
      parsed.pathname.startsWith('/images/') ||
      parsed.pathname === '/upload';
    return onCdn && filePath;
  } catch {
    return false;
  }
}

/**
 * The CDN path and image options to sign for a variant. Signing by path lets
 * the CDN resolve it against its own current host, so a file stored under an
 * earlier CDN host still signs. Resized variants read from `/images/{id}`.
 */
export function signTarget(
  url: string,
  variant: AssetVariant
): { path: string; params?: Record<string, string> } {
  const parsed = new URL(url);
  const file = FILE_PATH.exec(parsed.pathname);
  if (variant === 'original' || !file) {
    return { path: `${parsed.pathname}${parsed.search}` };
  }
  return { path: `/images/${file[1]}`, params: { w: VARIANT_WIDTH[variant] } };
}

/**
 * Every link signed within the same clock hour expires at the same moment, an
 * hour after that hour ends. Identical requests then get an identical link, so
 * the browser serves a repeat view from its cache instead of downloading again.
 */
export function windowExpiry(nowSeconds: number): number {
  return (Math.floor(nowSeconds / HOUR_SECONDS) + 2) * HOUR_SECONDS;
}

const nowSeconds = () => Math.floor(Date.now() / 1000);

// A signed link is a bearer credential, so it lives only in the in-memory query
// cache. It is never written to sessionStorage or localStorage, where any
// same-origin script could read it and it would outlive a logout or a business
// switch.
const signedAssetQuery = (url: string, variant: AssetVariant) => ({
  queryKey: ['signed-asset', variant, url],
  queryFn: async () => {
    const target = signTarget(url, variant);
    const signed = await signAssetUrl(target.path, {
      expiresAt: windowExpiry(nowSeconds()),
      params: target.params,
    });
    return signed;
  },
  staleTime: (query: { state: { data?: SignedAssetUrl } }) => {
    const expiresAt = query.state.data?.expiresAt;
    if (!expiresAt) return 0;
    return Math.max(0, (expiresAt - RESIGN_MARGIN_SECONDS - nowSeconds()) * 1000);
  },
  refetchInterval: (query: { state: { data?: SignedAssetUrl } }) => {
    const expiresAt = query.state.data?.expiresAt;
    if (!expiresAt) return false as const;
    return Math.max(60_000, (expiresAt - RESIGN_MARGIN_SECONDS - nowSeconds()) * 1000);
  },
  gcTime: HOUR_SECONDS * 1000,
  retry: 1,
});

export interface SignedAsset {
  /** The address to load, or null while a signed link is on its way. */
  url: string | null;
  loading: boolean;
}

/**
 * The address to load for a stored file: a signed link for our own CDN files,
 * the stored address for anything else. If signing fails the stored address is
 * returned, so a public file still loads.
 */
export function useSignedAssetUrl(
  url: string | null,
  variant: AssetVariant = 'original'
): SignedAsset {
  const onCdn = !!url && isCdnFileUrl(url);
  const query = useQuery({ ...signedAssetQuery(url ?? '', variant), enabled: onCdn });
  if (!url) return { url: null, loading: false };
  if (!onCdn) return { url, loading: false };
  if (query.data) return { url: query.data.url, loading: false };
  if (query.isError) return { url, loading: false };
  return { url: null, loading: true };
}

/** Resolves a stored file address to the original's link, for click handlers. */
export function useResolveAssetUrl(): (url: string) => Promise<string> {
  const queryClient = useQueryClient();
  return useCallback(
    async (url: string) => {
      if (!isCdnFileUrl(url)) return url;
      try {
        const signed = await queryClient.fetchQuery(signedAssetQuery(url, 'original'));
        return signed.url;
      } catch {
        return url;
      }
    },
    [queryClient]
  );
}
