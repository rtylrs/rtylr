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
});
