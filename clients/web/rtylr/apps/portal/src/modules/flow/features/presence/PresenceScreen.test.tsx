import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

// camelCase fixtures on purpose. The endpoint answers snake_case, but
// @voxire/api's response interceptor converts every key before a caller sees
// it, so this is the shape usePresence actually receives. The previous
// snake_case fixtures matched the schema rather than reality, which is why the
// suite stayed green while /presence showed its error state (#1067).
const presenceState = vi.hoisted(() => ({
  value: {
    data: {
      date: '2027-05-10',
      presence: [
        { employeeId: '1', userId: '10', firstName: 'Jane', lastName: 'Doe', status: 'onsite' },
        {
          employeeId: '2',
          userId: '20',
          firstName: 'Sam',
          lastName: 'Lee',
          status: 'on_leave',
        },
        {
          employeeId: '3',
          userId: '30',
          firstName: 'Alex',
          lastName: 'Kim',
          status: 'not_started',
        },
      ],
    },
    isLoading: false,
    isError: false,
  },
}));

const usePresenceMock = vi.hoisted(() => vi.fn());
vi.mock('./usePresence', () => ({
  usePresence: () => usePresenceMock(),
}));

import PresenceScreen from './PresenceScreen';

beforeEach(() => {
  vi.clearAllMocks();
  presenceState.value = {
    data: {
      date: '2027-05-10',
      presence: [
        { employeeId: '1', userId: '10', firstName: 'Jane', lastName: 'Doe', status: 'onsite' },
        {
          employeeId: '2',
          userId: '20',
          firstName: 'Sam',
          lastName: 'Lee',
          status: 'on_leave',
        },
        {
          employeeId: '3',
          userId: '30',
          firstName: 'Alex',
          lastName: 'Kim',
          status: 'not_started',
        },
      ],
    },
    isLoading: false,
    isError: false,
  };
  usePresenceMock.mockImplementation(() => presenceState.value);
});

describe('PresenceScreen', () => {
  it("shows every employee's resolved status", () => {
    render(<PresenceScreen />);
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('flow.time.mode.onsite')).toBeInTheDocument();
    expect(screen.getByText('Sam Lee')).toBeInTheDocument();
    expect(screen.getByText('flow.presence.onLeave')).toBeInTheDocument();
    expect(screen.getByText('Alex Kim')).toBeInTheDocument();
    expect(screen.getByText('flow.presence.notStarted')).toBeInTheDocument();
  });

  it('shows a no-results state for an empty roster', () => {
    presenceState.value = {
      data: { date: '2027-05-10', presence: [] },
      isLoading: false,
      isError: false,
    };
    render(<PresenceScreen />);
    expect(screen.getByText('flow.presence.noResults')).toBeInTheDocument();
  });

  it('shows a load-failed state, not a crash, when presence denies the caller', () => {
    presenceState.value = { data: undefined, isLoading: false, isError: true };
    render(<PresenceScreen />);
    expect(screen.getByText('flow.presence.loadFailedTitle')).toBeInTheDocument();
  });

  it('shows the load-failed state when a denied query is paused after failing', () => {
    presenceState.value = {
      data: undefined,
      isLoading: false,
      isError: false,
      isPaused: true,
      failureCount: 2,
      refetch: vi.fn(),
    };
    render(<PresenceScreen />);
    expect(screen.getByText('flow.presence.loadFailedTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.presence.noResults')).not.toBeInTheDocument();
  });

  it('does not claim an empty roster before any data has arrived', () => {
    presenceState.value = { data: undefined, isLoading: false, isError: false, refetch: vi.fn() };
    render(<PresenceScreen />);
    expect(screen.queryByText('flow.presence.noResults')).not.toBeInTheDocument();
  });
});

describe('PresenceScreen grouping (#1094)', () => {
  const withPresence = (presence: Record<string, unknown>[]): void => {
    presenceState.value = {
      data: { date: '2027-05-10', presence },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    };
  };

  const person = (id: string, status: string, first = 'A', last = 'B') => ({
    employeeId: id,
    userId: id,
    firstName: first,
    lastName: last,
    status,
  });

  it('groups people under working, on leave and not started', () => {
    withPresence([
      person('1', 'onsite', 'Jane', 'Doe'),
      person('2', 'on_leave', 'Sam', 'Lee'),
      person('3', 'not_started', 'Alex', 'Kim'),
    ]);
    render(<PresenceScreen />);

    const headings = screen.getAllByRole('heading', { level: 2 }).map((h) => h.textContent ?? '');
    expect(headings.some((h) => h.includes('flow.presence.working'))).toBe(true);
    expect(headings.some((h) => h.includes('flow.presence.onLeave'))).toBe(true);
    expect(headings.some((h) => h.includes('flow.presence.notStarted'))).toBe(true);
  });

  it('counts the people in each group', () => {
    withPresence([
      person('1', 'onsite', 'Jane', 'Doe'),
      person('2', 'remote', 'Ravi', 'Shah'),
      person('3', 'on_leave', 'Sam', 'Lee'),
    ]);
    render(<PresenceScreen />);

    const working = screen
      .getAllByRole('heading', { level: 2 })
      .find((h) => (h.textContent ?? '').includes('flow.presence.working'));
    expect(working?.textContent).toContain('2');
  });

  it('leaves out a group nobody is in, rather than showing an empty heading', () => {
    // An empty "On leave" heading reads as a section that failed to load.
    withPresence([person('1', 'onsite', 'Jane', 'Doe')]);
    render(<PresenceScreen />);

    const headings = screen.getAllByRole('heading', { level: 2 }).map((h) => h.textContent ?? '');
    expect(headings.some((h) => h.includes('flow.presence.onLeave'))).toBe(false);
  });

  it('keeps a row badge only where the status varies within the group', () => {
    // Under "On leave" the badge repeated the heading word for word on every
    // row; under "Working" it is what tells onsite from remote.
    withPresence([person('1', 'remote', 'Ravi', 'Shah'), person('2', 'on_leave', 'Sam', 'Lee')]);
    render(<PresenceScreen />);

    expect(screen.getByText('flow.time.mode.remote')).toBeInTheDocument();
    // Once, as the heading, not twice.
    expect(screen.getAllByText(/flow\.presence\.onLeave/)).toHaveLength(1);
  });

  it('draws a glyph beside every group heading', () => {
    // S-22 wants the status carried by more than colour and position.
    withPresence([person('1', 'onsite', 'Jane', 'Doe')]);
    const { container } = render(<PresenceScreen />);
    const heading = screen.getAllByRole('heading', { level: 2 })[0];
    expect(heading.querySelector('svg'), 'heading has no glyph').toBeTruthy();
    expect(container.querySelectorAll('svg').length).toBeGreaterThan(0);
  });
});
