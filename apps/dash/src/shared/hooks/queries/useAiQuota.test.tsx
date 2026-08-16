import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useAiQuotaMode, useUpdateAiQuotaMode } from './useAiQuota';

const mockGet = vi.fn();
const mockPut = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper(client: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('AI quota queries', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('updates the cached mode from PUT without refetching quota settings', async () => {
    mockGet.mockResolvedValue({ settings: { mode: 'pool' } });
    mockPut.mockResolvedValue({ settings: { mode: 'per_person' } });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(
      () => ({ mode: useAiQuotaMode(), update: useUpdateAiQuotaMode() }),
      { wrapper }
    );

    await waitFor(() => expect(result.current.mode.isSuccess).toBe(true));

    act(() => {
      result.current.update.mutate('per_person');
    });

    await waitFor(() => expect(result.current.update.isSuccess).toBe(true));

    expect(mockPut).toHaveBeenCalledWith('/hub/ai/v1/quota-settings', { mode: 'per_person' });
    expect(mockGet).toHaveBeenCalledTimes(1);
    expect(client.getQueryData(['ai-quota-settings'])).toBe('per_person');
  });
});
