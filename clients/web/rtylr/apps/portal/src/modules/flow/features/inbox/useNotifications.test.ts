// @vitest-environment happy-dom
import React from 'react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getNotifications: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}));

import { flowApi } from '../../shared/services/flowApi';
import { useNotifications } from './useNotifications';

function wrapper({ children }: { children: React.ReactNode }) {
  return React.createElement(
    QueryClientProvider,
    { client: new QueryClient({ defaultOptions: { queries: { retry: false } } }) },
    children
  );
}

describe('useNotifications', () => {
  beforeEach(() => vi.clearAllMocks());

  it('loads the complete history rather than presenting page one as the inbox', async () => {
    vi.mocked(flowApi.getNotifications)
      .mockResolvedValueOnce({
        notifications: [{ id: 'n-1' }],
        pagination: { page: 1, hasNext: true },
      } as never)
      .mockResolvedValueOnce({
        notifications: [{ id: 'n-2' }],
        pagination: { page: 2, hasNext: false },
      } as never);

    const { result } = renderHook(() => useNotifications(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(flowApi.getNotifications).toHaveBeenNthCalledWith(1, { page: 1, limit: 500 });
    expect(flowApi.getNotifications).toHaveBeenNthCalledWith(2, { page: 2, limit: 500 });
  });
});
