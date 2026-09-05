import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

type Notification = {
  id: string;
  read: boolean;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  body: string;
  app: string;
  createdAt: string;
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
  useMarkNotificationRead: () => ({ mutate: vi.fn() }),
  useMarkAllNotificationsRead: () => ({ mutate: vi.fn() }),
}));

import Notifications from './Notifications';

beforeEach(() => {
  vi.clearAllMocks();
  notificationsData.pages = undefined;
  notificationsData.isLoading = false;
});

describe('Notifications empty state', () => {
  it('shows the all-caught-up empty state with a refresh action on the "all" tab', () => {
    render(<Notifications />);

    expect(screen.getByText('notifications.emptyAllCaughtUpTitle')).toBeInTheDocument();
    expect(screen.getByText('notifications.emptyNoneDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches notifications when the refresh action is clicked on the "all" tab', () => {
    render(<Notifications />);

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(notificationsData.refetch).toHaveBeenCalledTimes(1);
  });

  it('shows a "view all" action instead of refresh on the "unread" tab when everything is read', () => {
    setPages([makeNotification({ read: true })]);
    render(<Notifications />);

    fireEvent.click(screen.getByRole('button', { name: 'unread' }));

    expect(screen.getByText('notifications.emptyAllCaughtUpTitle')).toBeInTheDocument();
    expect(screen.getByText('notifications.emptyNoUnreadDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'notifications.viewAll' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'common.refresh' })).not.toBeInTheDocument();
  });

  it('switches back to the "all" tab when "view all" is clicked from the unread empty state', () => {
    setPages([makeNotification({ read: true, title: 'Read receipt' })]);
    render(<Notifications />);

    fireEvent.click(screen.getByRole('button', { name: 'unread' }));
    expect(screen.getByText('notifications.emptyNoUnreadDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'notifications.viewAll' }));

    // Back on the "all" tab the read notification is shown instead of the empty state.
    expect(screen.queryByText('notifications.emptyNoUnreadDescription')).not.toBeInTheDocument();
    expect(screen.getByText('Read receipt')).toBeInTheDocument();
  });

  it('does not show the empty state when there are notifications', () => {
    setPages([makeNotification({ read: false })]);
    render(<Notifications />);

    expect(screen.queryByText('notifications.emptyAllCaughtUpTitle')).not.toBeInTheDocument();
  });
});
