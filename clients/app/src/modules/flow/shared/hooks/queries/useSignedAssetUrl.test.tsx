import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';

const signAssetUrl = vi.hoisted(() => vi.fn());
vi.mock('@voxire/api', () => ({ signAssetUrl }));
vi.mock('@voxire/config', () => ({ getCdnBaseUrl: () => 'https://api.voxire.com' }));

import { signTarget, useSignedAssetUrl, windowExpiry } from './useSignedAssetUrl';

const FILE = 'https://cdn.voxire.com/files/0f1e2d3c-4b5a-4678-9abc-def012345678';

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe('windowExpiry', () => {
  it('gives every request in the same hour the same expiry, at least an hour away', () => {
    const start = 1_800_000_000 - (1_800_000_000 % 3600);
    expect(windowExpiry(start)).toBe(start + 7200);
    expect(windowExpiry(start + 3599)).toBe(start + 7200);
    expect(windowExpiry(start + 3600)).toBe(start + 10800);
    expect(windowExpiry(start + 3599) - (start + 3599)).toBeGreaterThan(3600);
  });
});

describe('signTarget', () => {
  it('asks for a resized copy of a stored file, and the file itself for the original', () => {
    expect(signTarget(FILE, 'thumbnail')).toEqual({
      path: '/images/0f1e2d3c-4b5a-4678-9abc-def012345678',
      params: { w: '160' },
    });
    expect(signTarget(FILE, 'preview').params).toEqual({ w: '1600' });
    expect(signTarget(FILE, 'original')).toEqual({
      path: '/files/0f1e2d3c-4b5a-4678-9abc-def012345678',
    });
  });
});

describe('useSignedAssetUrl', () => {
  beforeEach(() => {
    sessionStorage.clear();
    signAssetUrl.mockReset();
  });

  it('reuses a link signed earlier in the tab instead of waiting on another request', async () => {
    // Reopening a task waited on a signing request before any image could load.
    const expiresAt = Math.floor(Date.now() / 1000) + 3600;
    signAssetUrl.mockResolvedValue({ url: `${FILE}?sig=first`, expiresAt });
    const first = renderHook(() => useSignedAssetUrl(FILE), { wrapper });
    await waitFor(() => expect(first.result.current.url).toBe(`${FILE}?sig=first`));
    expect(signAssetUrl).toHaveBeenCalledTimes(1);

    const second = renderHook(() => useSignedAssetUrl(FILE), { wrapper });
    expect(second.result.current.url).toBe(`${FILE}?sig=first`);
    expect(signAssetUrl).toHaveBeenCalledTimes(1);
  });

  it('signs again when the stored link is about to expire', async () => {
    const soon = Math.floor(Date.now() / 1000) + 60;
    sessionStorage.setItem(
      `flow-signed-asset:original:${FILE}`,
      JSON.stringify({ url: `${FILE}?sig=old`, expiresAt: soon })
    );
    signAssetUrl.mockResolvedValue({
      url: `${FILE}?sig=new`,
      expiresAt: Math.floor(Date.now() / 1000) + 3600,
    });

    const { result } = renderHook(() => useSignedAssetUrl(FILE), { wrapper });

    await waitFor(() => expect(result.current.url).toBe(`${FILE}?sig=new`));
  });
});
