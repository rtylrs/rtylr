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

import { useQuotes } from './useQuotes';
import { safe } from '../../shared/lib/api';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useQuotes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches quotes on mount with default pagination', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      quotes: [{ id: '1', client: 'Alpha' }],
      has_more: false,
    });
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/quotes', {
      params: { page: 1, page_size: 50 },
    });
  });

  it('accepts custom page and pageSize params', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({ quotes: [], has_more: false });
    const { result } = renderHook(() => useQuotes({ page: 2, pageSize: 20 }), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(safe.get).toHaveBeenCalledWith('/finance/quotes', {
      params: { page: 2, page_size: 20 },
    });
  });

  it('exposes hasMore from response', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      quotes: [{ id: '2', client: 'Beta' }],
      has_more: true,
    });
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.hasMore).toBe(true);
  });

  it('starts in loading state', () => {
    (safe.get as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    expect(result.current.isLoading).toBe(true);
  });

  it('coerces numeric quote id to string', async () => {
    (safe.get as ReturnType<typeof vi.fn>).mockResolvedValue({
      quotes: [{ id: 42, client: 'Gamma' }],
      has_more: false,
    });
    const { result } = renderHook(() => useQuotes(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.quotes[0]?.id).toBe('42');
  });
});
