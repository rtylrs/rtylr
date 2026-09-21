import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useUserStore: vi.fn(),
  useTimeOff: vi.fn(),
  approveMutate: vi.fn(),
  navigate: vi.fn(),
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: mocks.useUserStore,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  useNavigate: () => mocks.navigate,
}));

vi.mock('../../shared/hooks/useHrExtras', () => ({
  useTimeOff: mocks.useTimeOff,
  useApproveTimeOff: () => ({ mutate: mocks.approveMutate }),
}));

import TimeOff from './TimeOff';

describe('TimeOff empty state', () => {
  beforeEach(() => {
    mocks.useUserStore.mockReset();
    mocks.useTimeOff.mockReset();
    mocks.approveMutate.mockReset();
    mocks.navigate.mockReset();

    mocks.useUserStore.mockImplementation(
      (selector: (state: { user: { role: string } }) => unknown) =>
        selector({ user: { role: 'staff' } })
    );
  });

  it('shows the empty state with title and description when there are no requests', () => {
    mocks.useTimeOff.mockReturnValue({ data: [], isLoading: false });

    render(<TimeOff />);

    expect(screen.getByText('timeOff.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('timeOff.emptyDescription')).toBeInTheDocument();
  });

  it('navigates to the new time-off request page when the CTA is clicked', () => {
    mocks.useTimeOff.mockReturnValue({ data: [], isLoading: false });

    render(<TimeOff />);

    fireEvent.click(screen.getByRole('button', { name: 'timeOff.submitFirstOne' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/time-off/new' });
  });

  it('shows a retryable error instead of the empty state when the fetch fails', () => {
    const refetch = vi.fn();
    mocks.useTimeOff.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      isFetching: false,
      refetch,
    });

    render(<TimeOff />);

    expect(screen.queryByText('timeOff.emptyTitle')).not.toBeInTheDocument();
    expect(screen.getByRole('alert')).toHaveTextContent('common.loadError');

    fireEvent.click(screen.getByRole('button', { name: 'common.retry' }));

    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('does not render the empty state once requests exist', () => {
    mocks.useTimeOff.mockReturnValue({
      data: [
        {
          id: 1,
          staffName: 'Alex Chen',
          status: 'Pending',
          type: 'Vacation',
          startDate: '2026-01-01',
          endDate: '2026-01-02',
          reason: 'Trip',
        },
      ],
      isLoading: false,
    });

    render(<TimeOff />);

    expect(screen.queryByText('timeOff.emptyTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Alex Chen')).toBeInTheDocument();
  });
});
