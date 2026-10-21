// @vitest-environment happy-dom
import React from 'react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getProjectColumns: vi.fn(),
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
import { useProjectColumns } from './useColumns';

function wrapper({ children }: { children: React.ReactNode }) {
  return React.createElement(
    QueryClientProvider,
    { client: new QueryClient({ defaultOptions: { queries: { retry: false } } }) },
    children
  );
}

describe('useProjectColumns', () => {
  beforeEach(() => vi.clearAllMocks());

  it('loads every page so later columns are available to the board', async () => {
    vi.mocked(flowApi.getProjectColumns)
      .mockResolvedValueOnce({
        columns: [{ id: '1', name: 'To Do' }],
        pagination: { page: 1, hasNext: true },
      } as never)
      .mockResolvedValueOnce({
        columns: [{ id: '2', name: 'Done' }],
        pagination: { page: 2, hasNext: false },
      } as never);

    const { result } = renderHook(() => useProjectColumns('p-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(flowApi.getProjectColumns).toHaveBeenNthCalledWith(1, 'p-1', {
      page: 1,
      limit: 500,
    });
    expect(flowApi.getProjectColumns).toHaveBeenNthCalledWith(2, 'p-1', {
      page: 2,
      limit: 500,
    });
  });
});
