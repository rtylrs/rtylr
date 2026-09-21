// @vitest-environment happy-dom
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import Settings from './Settings';
import { initTestI18n } from '../../shared/testing/i18n';

const { useFinanceTeamMock, refetchMock } = vi.hoisted(() => ({
  useFinanceTeamMock: vi.fn(),
  refetchMock: vi.fn(),
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (state: { user: { role: string }; role: string }) => unknown) =>
    selector({ user: { role: 'owner' }, role: 'owner' }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

// i18next is deliberately not mocked so the assertions below also prove the
// locale keys resolve (see Quotes.test.tsx and rtylr#1269).
vi.mock('../../shared/hooks/queries/useFinanceTeam', () => ({
  useFinanceTeam: useFinanceTeamMock,
  useFinanceInviteMember: () => ({ mutate: vi.fn(), isPending: false }),
  useFinanceDeactivateMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useFinanceSubscription', () => ({
  useFinanceSubscription: () => ({ data: null }),
  centsToDisplay: (cents: number) => String(cents),
  computeSeatProration: () => ({
    daysRemaining: 1,
    daysInMonth: 30,
    proratedCents: 0,
    nextMonthName: 'Jan',
  }),
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ title }: { title: string }) => <div>{title}</div>,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

const openTeamTab = (): void => {
  fireEvent.click(screen.getByRole('button', { name: 'Team' }));
};

describe('Finance team tab', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useFinanceTeamMock.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      refetch: refetchMock,
    });
  });

  it('shows the member full name, not the email, when the contract returns full_name', () => {
    useFinanceTeamMock.mockReturnValue({
      data: [
        {
          id: '7',
          fullName: 'Jane Doe',
          email: 'jane@x.com',
          role: 'staff',
          apps: ['finance'],
          isActive: true,
        },
      ],
      isLoading: false,
      isError: false,
      refetch: refetchMock,
    });

    render(<Settings />);
    openTeamTab();

    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('J')).toBeInTheDocument();
    expect(screen.getByText('jane@x.com')).toBeInTheDocument();
    expect(screen.queryByText('No team members yet')).not.toBeInTheDocument();
  });

  it('shows a retryable error state instead of the empty state when the members query fails', () => {
    useFinanceTeamMock.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: refetchMock,
    });

    render(<Settings />);
    openTeamTab();

    expect(screen.getByRole('alert')).toHaveTextContent("Couldn't load team members");
    expect(screen.queryByText('No team members yet')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(refetchMock).toHaveBeenCalledTimes(1);
  });
});
