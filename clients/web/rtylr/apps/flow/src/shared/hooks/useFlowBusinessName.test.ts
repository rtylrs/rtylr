import { act, renderHook, waitFor } from '@testing-library/react';
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

import { useFlowBusinessName, useFlowBusinessId } from './useFlowBusinessName';
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

describe('useFlowBusinessName and useFlowBusinessId share one cache entry', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // The two hooks used the same queryKey with a different value each (name vs
  // stringified id). The first-mounted observer's queryFn decided the value
  // both hooks read, so the realtime client subscribed to `business:<name>`
  // and a refetch could make the shell title render the numeric id.
  it('keeps the name and id projections apart, before and after a refetch', async () => {
    (flowApi.getMyBusinesses as unknown as Mock).mockResolvedValue({
      businesses: [
        { id: 1, name: 'Old Business', isActive: false },
        { id: 42, name: 'Acme Inc', isActive: true },
      ],
    });

    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);

    // Name first, id second: the order FlowLayout mounts them in.
    const { result } = renderHook(
      () => ({ name: useFlowBusinessName(), id: useFlowBusinessId() }),
      { wrapper }
    );

    await waitFor(() => expect(result.current.id.data).toBe('42'));
    expect(result.current.name.data).toBe('Acme Inc');
    // One fetch serves both projections.
    expect(flowApi.getMyBusinesses).toHaveBeenCalledTimes(1);

    await act(async () => {
      await client.invalidateQueries();
    });

    expect(result.current.name.data).toBe('Acme Inc');
    expect(result.current.id.data).toBe('42');
  });
});
