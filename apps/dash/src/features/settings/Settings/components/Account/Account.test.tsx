import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('@voxire/auth', () => ({
  signOut: vi.fn(),
}));

// Stable reference: a mock that returns a fresh object each call changes the
// `useEffect(..., [user])` dependency identity on every render and loops forever.
const STABLE_USER = vi.hoisted(() => ({ fullName: 'Jane Doe', email: 'jane@rtylr.com' }));

vi.mock('../../../../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: STABLE_USER }),
}));

type LoginEntry = {
  id: string;
  email: string;
  success: boolean;
  createdAt: string;
  method: string;
  ip: string;
  userAgent: string;
  timeAgo: string;
  device: { kind: 'desktop'; label: string; description: string };
};

const loginHistoryState = vi.hoisted(() => ({
  data: undefined as
    | { entries: LoginEntry[]; total: number; page: number; pageSize: number; hasMore: boolean }
    | undefined,
  isLoading: false,
  isError: false,
  isFetching: false,
  refetch: vi.fn(),
}));

vi.mock('@tanstack/react-query', () => ({
  keepPreviousData: 'keepPreviousData',
  useQuery: () => loginHistoryState,
  useMutation: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
}));

import Account from './Account';

beforeEach(() => {
  vi.clearAllMocks();
  loginHistoryState.data = { entries: [], total: 0, page: 1, pageSize: 20, hasMore: false };
  loginHistoryState.isLoading = false;
  loginHistoryState.isError = false;
  loginHistoryState.isFetching = false;
});

describe('Account: login history empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there is no login history', () => {
    render(<Account />);

    expect(screen.getByText('settings.account.noLoginActivityTitle')).toBeInTheDocument();
    expect(screen.getByText('settings.account.noLoginActivityDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches the login history when the refresh call-to-action is clicked', () => {
    render(<Account />);

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(loginHistoryState.refetch).toHaveBeenCalledTimes(1);
  });

  it('shows a loading message instead of the empty state while loading', () => {
    loginHistoryState.isLoading = true;
    render(<Account />);

    expect(screen.getByText('Loading login history…')).toBeInTheDocument();
    expect(screen.queryByText('settings.account.noLoginActivityTitle')).not.toBeInTheDocument();
  });

  it('does not show the empty state when there is login history', () => {
    loginHistoryState.data = {
      entries: [
        {
          id: '1',
          email: 'jane@rtylr.com',
          success: true,
          createdAt: new Date().toISOString(),
          method: 'Password',
          ip: '1.2.3.4',
          userAgent: '',
          timeAgo: 'just now',
          device: { kind: 'desktop', label: 'Chrome on macOS', description: 'Desktop' },
        },
      ],
      total: 1,
      page: 1,
      pageSize: 20,
      hasMore: false,
    };
    render(<Account />);

    expect(screen.queryByText('settings.account.noLoginActivityTitle')).not.toBeInTheDocument();
  });
});
