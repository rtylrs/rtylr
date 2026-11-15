import { signAssetUrl } from '@voxire/api';
import { getCdnBaseUrl } from '@voxire/config';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useCallback } from 'react';

// Signed links last an hour. Re-sign a little before that so an image that is
// still on screen never points at an expired link.
const SIGNED_LINK_SECONDS = 60 * 60;
const RESIGN_AFTER_MS = 50 * 60 * 1000;

const KNOWN_CDN_HOSTS = new Set([
  'cdn.voxire.com',
  'cdn.voxire.dev',
  'api.voxire.com',
  'api.voxire.dev',
]);

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

// The CDN signs a path against its own current host. Sending the full stored
// address would be refused if the file was stored under an earlier CDN host.
function cdnPath(url: string): string {
  const parsed = new URL(url);
  return `${parsed.pathname}${parsed.search}`;
}

const signedAssetQuery = (url: string) => ({
  queryKey: ['signed-asset', url],
  queryFn: () => signAssetUrl(cdnPath(url), { expiresIn: SIGNED_LINK_SECONDS }),
  staleTime: RESIGN_AFTER_MS,
  gcTime: RESIGN_AFTER_MS,
  refetchInterval: RESIGN_AFTER_MS,
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
export function useSignedAssetUrl(url: string | null): SignedAsset {
  const onCdn = !!url && isCdnFileUrl(url);
  const query = useQuery({ ...signedAssetQuery(url ?? ''), enabled: onCdn });
  if (!url) return { url: null, loading: false };
  if (!onCdn) return { url, loading: false };
  if (query.data) return { url: query.data.url, loading: false };
  if (query.isError) return { url, loading: false };
  return { url: null, loading: true };
}

/** Resolves a stored file address to one that opens, for click handlers. */
export function useResolveAssetUrl(): (url: string) => Promise<string> {
  const queryClient = useQueryClient();
  return useCallback(
    async (url: string) => {
      if (!isCdnFileUrl(url)) return url;
      try {
        const signed = await queryClient.fetchQuery(signedAssetQuery(url));
        return signed.url;
      } catch {
        return url;
      }
    },
    [queryClient]
  );
}
