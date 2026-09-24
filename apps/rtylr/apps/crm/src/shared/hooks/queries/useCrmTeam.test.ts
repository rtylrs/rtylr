import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  listTeam: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    listTeam: mocks.listTeam,
    inviteUser: vi.fn(),
    removeMember: vi.fn(),
  }),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import { useCrmTeam } from './useCrmTeam';

function createWrapper(): React.FC<{ children: React.ReactNode }> {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mocks.listTeam.mockReset();
});

describe('useCrmTeam', () => {
  it('loads every page instead of treating the first 50 members as the team', async () => {
    const members = Array.from({ length: 51 }, (_, index) => ({
      id: String(index + 1),
      email: `member-${index + 1}@example.com`,
    }));
    mocks.listTeam.mockImplementation(async ({ page }: { page?: number }) => ({
      users: page === 1 ? members.slice(0, 50) : members.slice(50),
      total: members.length,
    }));

    const { result } = renderHook(() => useCrmTeam(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(51);
    expect(mocks.listTeam).toHaveBeenNthCalledWith(1, { page: 1, pageSize: 50 });
    expect(mocks.listTeam).toHaveBeenNthCalledWith(2, { page: 2, pageSize: 50 });
  });
});
