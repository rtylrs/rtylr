// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/lib/api', () => ({
  safe: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val: unknown) => Promise.resolve(val)),
}));

import { useBills } from './useBills';
import { safe } from '../../shared/lib/api';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useBills', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches bills on mount', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      bills: [{ id: '1', supplier: 'ACME', amount: 500 }],
    });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/bills');
    expect(result.current.data).toHaveLength(1);
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('returns empty array when bills key is empty', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ bills: [] });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([]);
  });

  it('coerces numeric bill id to string', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      bills: [{ id: 99, supplier: 'Vendor' }],
    });
    const { result } = renderHook(() => useBills(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.[0]?.id).toBe('99');
  });
});
