import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const presenceState = vi.hoisted(() => ({
  value: {
    data: {
      date: '2027-05-10',
      presence: [
        { employee_id: '1', user_id: '10', first_name: 'Jane', last_name: 'Doe', status: 'onsite' },
        {
          employee_id: '2',
          user_id: '20',
          first_name: 'Sam',
          last_name: 'Lee',
          status: 'on_leave',
        },
        {
          employee_id: '3',
          user_id: '30',
          first_name: 'Alex',
          last_name: 'Kim',
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
        { employee_id: '1', user_id: '10', first_name: 'Jane', last_name: 'Doe', status: 'onsite' },
        {
          employee_id: '2',
          user_id: '20',
          first_name: 'Sam',
          last_name: 'Lee',
          status: 'on_leave',
        },
        {
          employee_id: '3',
          user_id: '30',
          first_name: 'Alex',
          last_name: 'Kim',
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
