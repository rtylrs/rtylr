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

const DEFAULT_CDN_BASE_URL = 'https://cdn.voxire.com';

const getCdnBaseUrl = (override?: string): string => {
  const env = (import.meta as ImportMeta & { env?: Record<string, string | undefined> }).env;
  return (override || env?.VITE_CDN_URL || DEFAULT_CDN_BASE_URL).replace(/\/+$/, '');
};

const getToken = (): string | null => {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  const token = match ? decodeURIComponent(match[1]) : null;
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
