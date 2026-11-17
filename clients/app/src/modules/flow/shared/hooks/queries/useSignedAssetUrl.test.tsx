import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';

const signAssetUrl = vi.hoisted(() => vi.fn());
vi.mock('@voxire/api', () => ({ signAssetUrl }));
vi.mock('@voxire/config', () => ({ getCdnBaseUrl: () => 'https://api.voxire.com' }));

import { signTarget, useSignedAssetUrl, windowExpiry } from './useSignedAssetUrl';

const FILE = 'https://cdn.voxire.com/files/0f1e2d3c-4b5a-4678-9abc-def012345678';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
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
    localStorage.clear();
    signAssetUrl.mockReset();
  });

  it('reuses a link signed earlier in the session instead of waiting on another request', async () => {
    // Reopening a task waited on a signing request before any image could load.
    const wrapper = makeWrapper();
    const expiresAt = Math.floor(Date.now() / 1000) + 3600;
    signAssetUrl.mockResolvedValue({ url: `${FILE}?sig=first`, expiresAt });
    const first = renderHook(() => useSignedAssetUrl(FILE), { wrapper });
    await waitFor(() => expect(first.result.current.url).toBe(`${FILE}?sig=first`));
    expect(signAssetUrl).toHaveBeenCalledTimes(1);

    const second = renderHook(() => useSignedAssetUrl(FILE), { wrapper });
    expect(second.result.current.url).toBe(`${FILE}?sig=first`);
    expect(signAssetUrl).toHaveBeenCalledTimes(1);
  });

  it('never writes a signed link to web storage', async () => {
    // A signed link is a bearer credential. In web storage any same-origin
    // script could read it, and it would outlive a logout or business switch.
    signAssetUrl.mockResolvedValue({
      url: `${FILE}?sig=secret`,
      expiresAt: Math.floor(Date.now() / 1000) + 3600,
    });
    const { result } = renderHook(() => useSignedAssetUrl(FILE), { wrapper: makeWrapper() });
    await waitFor(() => expect(result.current.url).toBe(`${FILE}?sig=secret`));

    expect(sessionStorage.length).toBe(0);
    expect(localStorage.length).toBe(0);
  });

  it('signs again in a fresh session instead of reusing an earlier link', async () => {
    // Logout and business switch reload the page, which starts a new query cache.
    signAssetUrl.mockResolvedValueOnce({
      url: `${FILE}?sig=old`,
      expiresAt: Math.floor(Date.now() / 1000) + 3600,
    });
    const before = renderHook(() => useSignedAssetUrl(FILE), { wrapper: makeWrapper() });
    await waitFor(() => expect(before.result.current.url).toBe(`${FILE}?sig=old`));

    signAssetUrl.mockResolvedValueOnce({
      url: `${FILE}?sig=new`,
      expiresAt: Math.floor(Date.now() / 1000) + 3600,
    });
    const after = renderHook(() => useSignedAssetUrl(FILE), { wrapper: makeWrapper() });
    expect(after.result.current.url).toBeNull();
    await waitFor(() => expect(after.result.current.url).toBe(`${FILE}?sig=new`));
    expect(signAssetUrl).toHaveBeenCalledTimes(2);
  });
});
