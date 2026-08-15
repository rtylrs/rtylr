import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}));

const userState = vi.hoisted(() => ({ value: { id: 'u1' } }));
vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (s: { user: { id: string } | undefined }) => unknown) =>
    selector({ user: userState.value }),
}));

vi.mock('../../shared/stores/timeSettingsStore', () => ({
  useTimeSettingsStore: (selector: (s: { weekStart: 'monday' | 'sunday' }) => unknown) =>
    selector({ weekStart: 'monday' }),
}));

const staffState = vi.hoisted(() => ({
  value: {
    data: {
      id: '9',
      userId: 'u1',
      firstName: 'Jane',
      lastName: 'Doe',
      role: 'Engineer',
      departmentName: 'Engineering',
      employmentType: 'Full-time',
      status: 'active',
      startDate: '2024-01-15',
      email: 'jane@example.com',
      phone: '+1-555-0100',
      address: '',
      emergencyContactName: '',
      emergencyContactPhone: '',
    } as Record<string, unknown> | null,
    isLoading: false,
  },
}));
const detailState = vi.hoisted(() => ({
  value: { data: undefined as { bio?: string } | undefined },
}));

vi.mock('./useProfile', () => ({
  useMyStaffRecord: () => staffState.value,
  useMyProfileDetail: () => detailState.value,
}));

const weekDaysState = vi.hoisted(() => ({
  value: { data: { workDays: [{ date: '2026-08-11', loggedMinutes: 120 }] } },
}));
vi.mock('../time/useWorkDay', () => ({
  todayLocalDate: () => '2026-08-16',
  useWorkDaysRange: () => weekDaysState.value,
}));

import ProfileScreen from './ProfileScreen';

beforeEach(() => {
  vi.clearAllMocks();
  userState.value = { id: 'u1' };
  staffState.value = {
    data: {
      id: '9',
      userId: 'u1',
      firstName: 'Jane',
      lastName: 'Doe',
      role: 'Engineer',
      departmentName: 'Engineering',
      employmentType: 'Full-time',
      status: 'active',
      startDate: '2024-01-15',
      email: 'jane@example.com',
      phone: '+1-555-0100',
      address: '',
      emergencyContactName: '',
      emergencyContactPhone: '',
    },
    isLoading: false,
  };
  detailState.value = { data: undefined };
  weekDaysState.value = { data: { workDays: [{ date: '2026-08-11', loggedMinutes: 120 }] } };
});

describe('ProfileScreen: HR record', () => {
  it("shows the caller's own HR record fields", () => {
    render(<ProfileScreen />);
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.getByText('Engineer')).toBeInTheDocument();
    expect(screen.getByText('Engineering')).toBeInTheDocument();
    expect(screen.getByText('jane@example.com')).toBeInTheDocument();
  });

  it('omits a field the backend did not return, rather than showing it blank', () => {
    render(<ProfileScreen />);
    expect(screen.queryByText('flow.profile.address')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.profile.emergencyContact')).not.toBeInTheDocument();
  });

  it('shows an address once the backend returns one', () => {
    staffState.value = {
      data: { ...(staffState.value.data as Record<string, unknown>), address: '12 Cedar St' },
      isLoading: false,
    };
    render(<ProfileScreen />);
    expect(screen.getByText('12 Cedar St')).toBeInTheDocument();
  });

  it('shows an empty state when no HR record is linked to the account', () => {
    staffState.value = { data: null, isLoading: false };
    render(<ProfileScreen />);
    expect(screen.getByText('flow.profile.noRecordTitle')).toBeInTheDocument();
  });

  it('shows bio once the profile-detail endpoint returns one', () => {
    detailState.value = { data: { bio: 'Builds things.' } };
    render(<ProfileScreen />);
    expect(screen.getByText('Builds things.')).toBeInTheDocument();
  });
});

describe('ProfileScreen: Flow activity', () => {
  it("shows the week's logged hours and a link back to My day", () => {
    render(<ProfileScreen />);
    expect(screen.getByText(/flow\.profile\.hoursLoggedThisWeek/)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'flow.profile.viewMyDay' })).toHaveAttribute(
      'href',
      '/time'
    );
  });
});
