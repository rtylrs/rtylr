import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const navigate = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigate }));

const effectiveAccess = vi.hoisted(() => ({ value: undefined as unknown }));
vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: unknown) => unknown) =>
    selector({ user: { effectiveAccess: effectiveAccess.value } }),
}));
vi.mock('@voxire/permissions', () => ({
  canEffectiveAccess: (access: unknown, permission: string) =>
    Array.isArray(access) && access.includes(permission),
}));

interface DirectoryQueryState {
  data: Array<Record<string, unknown>> | undefined;
  isLoading: boolean;
  isPending?: boolean;
  isError: boolean;
  isPaused?: boolean;
  failureCount?: number;
  error?: unknown;
  departmentsLoaded?: boolean;
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
vi.mock('./useDirectory', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./useDirectory')>()),
  useDirectory: (search: string | undefined) => useDirectoryMock(search),
}));

import { ApiError } from '@voxire/voxie/hr/staff/v1/client';
import DirectoryScreen from './DirectoryScreen';

const forbidden = (code: string) => new ApiError(403, JSON.stringify({ message: 'denied', code }));

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
  effectiveAccess.value = undefined;
  useDirectoryMock.mockImplementation(() => {
    const { data, departmentsLoaded = true, ...rest } = directoryState.value;
    return { ...rest, data: data && { members: data, departmentsLoaded } };
  });
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

  it('shows a load-failed state with retry for a non-permission failure', () => {
    directoryState.value = { data: undefined, isLoading: false, isError: true };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.loadFailedTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.directory.retryCta')).toBeInTheDocument();
  });

  it('says the directory is off, not that it failed, when the business turned it off', () => {
    directoryState.value = {
      data: undefined,
      isLoading: false,
      isError: true,
      error: forbidden('directory_disabled'),
    };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.disabledTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.directory.disabledDescription')).toBeInTheDocument();
    expect(screen.queryByText('flow.directory.loadFailedTitle')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.directory.openHrSettingsCta')).not.toBeInTheDocument();
  });

  it('offers HR settings to an admin when the directory is off', () => {
    effectiveAccess.value = ['hub:business:settings:manage'];
    directoryState.value = {
      data: undefined,
      isLoading: false,
      isError: true,
      error: forbidden('directory_disabled'),
    };
    render(<DirectoryScreen />);
    fireEvent.click(screen.getByText('flow.directory.openHrSettingsCta'));
    expect(navigate).toHaveBeenCalledWith({ to: '/settings/hr' });
  });

  it('says the role lacks access, with no pointless retry, when view_directory is missing', () => {
    directoryState.value = {
      data: undefined,
      isLoading: false,
      isError: true,
      error: forbidden('directory_permission_missing'),
    };
    render(<DirectoryScreen />);
    expect(screen.getByText('flow.directory.noAccessTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.directory.retryCta')).not.toBeInTheDocument();
  });

  it('still lists everyone, ungrouped and labelled, when departments fail to load', () => {
    directoryState.value = { ...directoryState.value, departmentsLoaded: false };
    render(<DirectoryScreen />);
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('Sam Lee')).toBeInTheDocument();
    expect(screen.getByText('flow.directory.departmentsUnavailable')).toBeInTheDocument();
    expect(screen.queryByText('flow.directory.noDepartment')).not.toBeInTheDocument();
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

  it('shows a loading skeleton for a query paused before its first attempt', () => {
    directoryState.value = {
      data: undefined,
      isLoading: false,
      isPending: true,
      isError: false,
      isPaused: true,
      failureCount: 0,
    };
    const { container } = render(<DirectoryScreen />);
    expect(container.querySelector('.flow-shimmer')).not.toBeNull();
    expect(screen.queryByText('flow.directory.noMembersYet')).not.toBeInTheDocument();
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
