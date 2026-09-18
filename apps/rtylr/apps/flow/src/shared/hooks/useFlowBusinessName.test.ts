import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import React from 'react';

vi.mock('../services/flowApi', () => ({
  flowApi: {
    getMyBusinesses: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useFlowBusinessName } from './useFlowBusinessName';
import { flowApi } from '../services/flowApi';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useFlowBusinessName', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('prefers the business flagged as active', async () => {
    (flowApi.getMyBusinesses as unknown as Mock).mockResolvedValue({
      businesses: [
        { id: 1, name: 'Old Business', isActive: false },
        { id: 2, name: 'Current Business', isActive: true },
      ],
    });
    const { result } = renderHook(() => useFlowBusinessName(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe('Current Business');
  });

  it('falls back to the first business when none is flagged active', async () => {
    (flowApi.getMyBusinesses as unknown as Mock).mockResolvedValue({
      businesses: [
        { id: 1, name: 'First Business' },
        { id: 2, name: 'Second Business' },
      ],
    });
    const { result } = renderHook(() => useFlowBusinessName(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe('First Business');
  });

  it('resolves null (not undefined) when the caller has no businesses', async () => {
    // TanStack Query v5 throws if a queryFn resolves to `undefined` - this
    // guards against that regression.
    (flowApi.getMyBusinesses as unknown as Mock).mockResolvedValue({ businesses: [] });
    const { result } = renderHook(() => useFlowBusinessName(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
  });
});
