import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

interface DirectoryQueryState {
  data: Array<Record<string, unknown>> | undefined;
  isLoading: boolean;
  isError: boolean;
  isPaused?: boolean;
  failureCount?: number;
}

const directoryState = vi.hoisted(() => ({
  value: {
    data: [
      {
        id: '1',
        firstName: 'Jane',
        lastName: 'Doe',
        role: 'Engineer',
        departmentName: 'Engineering',
        email: 'jane@example.com',
        phone: '+1-555-0100',
      },
      {
        id: '2',
        firstName: 'Sam',
        lastName: 'Lee',
        role: 'Designer',
        departmentName: 'Product',
        email: 'sam@example.com',
        phone: '+1-555-0101',
      },
    ],
    isLoading: false,
    isError: false,
  } as DirectoryQueryState,
}));

const useDirectoryMock = vi.hoisted(() => vi.fn());
vi.mock('./useDirectory', () => ({
  useDirectory: (search: string | undefined) => useDirectoryMock(search),
}));

import DirectoryScreen from './DirectoryScreen';

beforeEach(() => {
  vi.clearAllMocks();
  directoryState.value = {
    data: [
      {
        id: '1',
        firstName: 'Jane',
        lastName: 'Doe',
        role: 'Engineer',
        departmentName: 'Engineering',
        email: 'jane@example.com',
        phone: '+1-555-0100',
      },
      {
        id: '2',
        firstName: 'Sam',
        lastName: 'Lee',
        role: 'Designer',
        departmentName: 'Product',
        email: 'sam@example.com',
        phone: '+1-555-0101',
      },
    ],
    isLoading: false,
    isError: false,
  };
  useDirectoryMock.mockImplementation(() => directoryState.value);
});

describe('DirectoryScreen', () => {
  it('renders every roster member returned', () => {
    render(<DirectoryScreen />);
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('Sam Lee')).toBeInTheDocument();
  });

  it('re-queries with the typed search term', () => {
    render(<DirectoryScreen />);
    fireEvent.change(screen.getByLabelText('flow.directory.searchPlaceholder'), {
      target: { value: 'jane' },
    });
    expect(useDirectoryMock).toHaveBeenLastCalledWith('jane');
  });

  it('shows a true-empty state for a zero-member workspace with no search typed', () => {
    directoryState.value = { data: [], isLoading: false, isError: false };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.noMembersYet')).toBeInTheDocument();
    expect(screen.queryByText('flow.directory.noResults')).not.toBeInTheDocument();
  });

  it('shows a filtered no-results state once a search term matches nothing', () => {
    directoryState.value = { data: [], isLoading: false, isError: false };
    render(<DirectoryScreen />);
    fireEvent.change(screen.getByLabelText('flow.directory.searchPlaceholder'), {
      target: { value: 'nobody' },
    });
    expect(screen.getByText('flow.directory.noResults')).toBeInTheDocument();
    expect(screen.queryByText('flow.directory.noMembersYet')).not.toBeInTheDocument();
  });

  it('shows a load-failed state, not a crash, when listStaff denies the caller', () => {
    directoryState.value = { data: undefined, isLoading: false, isError: true };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.loadFailedTitle')).toBeInTheDocument();
  });

  it('shows the load-failed state for a query stuck paused after a failed attempt, not the empty state', () => {
    directoryState.value = {
      data: undefined,
      isLoading: false,
      isError: false,
      isPaused: true,
      failureCount: 1,
    };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.loadFailedTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.directory.noMembersYet')).not.toBeInTheDocument();
  });

  it('does not treat a query merely paused before its first attempt as failed', () => {
    directoryState.value = {
      data: undefined,
      isLoading: false,
      isError: false,
      isPaused: true,
      failureCount: 0,
    };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.noMembersYet')).toBeInTheDocument();
    expect(screen.queryByText('flow.directory.loadFailedTitle')).not.toBeInTheDocument();
  });
});

describe('DirectoryScreen groups by department (#1094)', () => {
  it('renders a heading and a count for each department', () => {
    directoryState.value = {
      data: [
        { id: '1', firstName: 'Ada', lastName: 'A', email: 'a@x.com', departmentName: 'Site' },
        { id: '2', firstName: 'Bea', lastName: 'B', email: 'b@x.com', departmentName: 'Site' },
        { id: '3', firstName: 'Cy', lastName: 'C', email: 'c@x.com', departmentName: 'Office' },
      ],
      isLoading: false,
      isError: false,
    } as never;
    render(<DirectoryScreen />);

    expect(screen.getByText('Office')).toBeInTheDocument();
    expect(screen.getByText('Site')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('1')).toBeInTheDocument();
  });

  it('names the group for people with no department', () => {
    directoryState.value = {
      data: [{ id: '1', firstName: 'Ada', lastName: 'A', email: 'a@x.com' }],
      isLoading: false,
      isError: false,
    } as never;
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.noDepartment')).toBeInTheDocument();
  });

  it('does not repeat the department on every row', () => {
    // The department is the heading now. It used to be appended to the role
    // on each line, so a department of eight people printed its name eight
    // times.
    directoryState.value = {
      data: [
        {
          id: '1',
          firstName: 'Ada',
          lastName: 'A',
          email: 'a@x.com',
          role: 'Foreman',
          departmentName: 'Site',
        },
      ],
      isLoading: false,
      isError: false,
    } as never;
    render(<DirectoryScreen />);

    expect(screen.getAllByText('Site')).toHaveLength(1);
    expect(screen.getByText('Foreman')).toBeInTheDocument();
    expect(screen.queryByText('Foreman · Site')).not.toBeInTheDocument();
  });
});
