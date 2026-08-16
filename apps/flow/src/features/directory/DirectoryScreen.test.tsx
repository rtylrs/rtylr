import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

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
    ] as Array<Record<string, unknown>> | undefined,
    isLoading: false,
    isError: false,
  },
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

  it('shows a no-results state for an empty roster', () => {
    directoryState.value = { data: [], isLoading: false, isError: false };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.noResults')).toBeInTheDocument();
  });

  it('shows a load-failed state, not a crash, when listStaff denies the caller', () => {
    directoryState.value = { data: undefined, isLoading: false, isError: true };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.loadFailedTitle')).toBeInTheDocument();
  });
});
