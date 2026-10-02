import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

type Notification = {
  id: string;
  read: boolean;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  body: string;
  app: string;
  createdAt: string;
  link?: string;
};

const makeNotification = (overrides: Partial<Notification>): Notification => ({
  id: '1',
  read: false,
  type: 'info',
  title: 'New order',
  body: 'You have a new order',
  app: 'pos',
  createdAt: new Date().toISOString(),
  ...overrides,
});

const navigateMock = vi.hoisted(() => vi.fn());
const markReadMock = vi.hoisted(() => vi.fn());
const notificationsData = vi.hoisted(() => ({
  pages: undefined as { pages: { notifications: Notification[] }[] } | undefined,
  isLoading: false,
  refetch: vi.fn(),
}));

const setPages = (list: Notification[]) => {
  notificationsData.pages = { pages: [{ notifications: list }] };
};

vi.mock('../../shared/hooks/queries/useNotifications', () => ({
  useNotifications: () => ({ data: [] }),
  useNotificationPages: () => ({
    data: notificationsData.pages,
    isLoading: notificationsData.isLoading,
    fetchNextPage: vi.fn(),
    hasNextPage: false,
    isFetchingNextPage: false,
    refetch: notificationsData.refetch,
  }),
  useMarkNotificationRead: () => ({ mutate: markReadMock }),
  useMarkAllNotificationsRead: () => ({ mutate: vi.fn() }),
}));

import Notifications, { resolveNotificationTarget } from './Notifications';

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  notificationsData.pages = undefined;
  notificationsData.isLoading = false;
});

describe('notification link safety', () => {
  it('accepts an absolute Dash URL and keeps its internal path', () => {
    expect(
      resolveNotificationTarget(
        'https://dash.example.test/team?from=notification#pending',
        'https://app.example.test',
        'https://dash.example.test'
      )
    ).toMatchObject({
      path: '/team?from=notification#pending',
      isCurrentOrigin: false,
    });
  });
});

describe('Notifications empty state', () => {
  it('shows the all-caught-up empty state with a refresh action on the "all" tab', () => {
    render(<Notifications />);

    expect(screen.getByText(i18next.t('notifications.emptyAllCaughtUpTitle'))).toBeInTheDocument();
    expect(screen.getByText(i18next.t('notifications.emptyNoneDescription'))).toBeInTheDocument();
    expect(screen.getByRole('button', { name: i18next.t('common.refresh') })).toBeInTheDocument();
  });

  it('refetches notifications when the refresh action is clicked on the "all" tab', () => {
    render(<Notifications />);

    fireEvent.click(screen.getByRole('button', { name: i18next.t('common.refresh') }));

    expect(notificationsData.refetch).toHaveBeenCalledTimes(1);
  });

  it('shows a "view all" action instead of refresh on the "unread" tab when everything is read', () => {
    setPages([makeNotification({ read: true })]);
    render(<Notifications />);

    fireEvent.click(screen.getByRole('button', { name: 'Unread' }));

    expect(screen.getByText(i18next.t('notifications.emptyAllCaughtUpTitle'))).toBeInTheDocument();
    expect(
      screen.getByText(i18next.t('notifications.emptyNoUnreadDescription'))
    ).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: i18next.t('notifications.viewAll') })
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: i18next.t('common.refresh') })
    ).not.toBeInTheDocument();
  });

  it('switches back to the "all" tab when "view all" is clicked from the unread empty state', () => {
    setPages([makeNotification({ read: true, title: 'Read receipt' })]);
    render(<Notifications />);

    fireEvent.click(screen.getByRole('button', { name: 'Unread' }));
    expect(
      screen.getByText(i18next.t('notifications.emptyNoUnreadDescription'))
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: i18next.t('notifications.viewAll') }));

    // Back on the "all" tab the read notification is shown instead of the empty state.
    expect(
      screen.queryByText(i18next.t('notifications.emptyNoUnreadDescription'))
    ).not.toBeInTheDocument();
    expect(screen.getByText('Read receipt')).toBeInTheDocument();
  });

  it('does not show the empty state when there are notifications', () => {
    setPages([makeNotification({ read: false })]);
    render(<Notifications />);

    expect(
      screen.queryByText(i18next.t('notifications.emptyAllCaughtUpTitle'))
    ).not.toBeInTheDocument();
  });

  it('marks a linked notification read and navigates to its internal Hub route', () => {
    setPages([makeNotification({ link: '/team' })]);
    render(<Notifications />);

    fireEvent.click(screen.getByText('New order'));

    expect(markReadMock).toHaveBeenCalledWith('1');
    expect(navigateMock).toHaveBeenCalledWith({ to: '/team' });
  });

  it('does not navigate to an unsafe notification scheme', () => {
    setPages([makeNotification({ link: 'javascript:alert(1)' })]);
    render(<Notifications />);

    fireEvent.click(screen.getByText('New order'));

    expect(markReadMock).toHaveBeenCalledWith('1');
    expect(navigateMock).not.toHaveBeenCalled();
  });
});
