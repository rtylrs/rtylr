// @vitest-environment happy-dom
import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockClient = vi.hoisted(() => ({
  listProjects: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mockClient,
}));
vi.mock('@voxire/voxie/flow/project/v2/client', () => ({
  ProjectServiceClient: {},
}));

import { useWorkProjects } from './useWorkProjects';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
};

describe('useWorkProjects', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('lists active shared projects by name and leaves out personal buckets', async () => {
    mockClient.listProjects.mockResolvedValue({
      projects: [
        { id: '3', name: 'Website', color: '#f00' },
        { id: '1', name: 'My Tasks', isPersonal: true },
        { id: '2', name: 'App redesign' },
      ],
    });

    const { result } = renderHook(() => useWorkProjects(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockClient.listProjects).toHaveBeenCalledWith(
      expect.objectContaining({ status: 'active', limit: 100 })
    );
    expect(result.current.data).toEqual([
      { id: '2', name: 'App redesign', color: undefined },
      { id: '3', name: 'Website', color: '#f00' },
    ]);
  });

  it('degrades to null when the caller cannot read Work projects', async () => {
    mockClient.listProjects.mockRejectedValue(new Error('403'));

    const { result } = renderHook(() => useWorkProjects(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
  });
});
