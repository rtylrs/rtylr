import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiNotification } from '../../shared/services/flowApi';

vi.mock('@voxire/i18n', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const notificationsState = vi.hoisted(() => ({
  data: [] as ApiNotification[],
  isLoading: false,
  refetch: vi.fn(),
}));
const markReadMutate = vi.hoisted(() => vi.fn());
const markAllMutate = vi.hoisted(() => vi.fn());

vi.mock('./useNotifications', () => ({
  useNotifications: () => notificationsState,
  useMarkNotificationRead: () => ({ mutate: markReadMutate, isPending: false }),
  useMarkAllNotificationsRead: () => ({ mutate: markAllMutate, isPending: false }),
}));

import Inbox from './Inbox';

function makeNotification(overrides: Partial<ApiNotification> = {}): ApiNotification {
  return {
    id: 'n1',
    type: 'assigned',
    read: false,
    title: 'You were assigned a task',
    createdAt: '2026-08-20T09:00:00Z',
    ...overrides,
  } as ApiNotification;
}

beforeEach(() => {
  vi.clearAllMocks();
  notificationsState.data = [];
  notificationsState.isLoading = false;
});

describe('Inbox empty state', () => {
  it('shows the empty title/description with a Refresh CTA when there are no notifications at all', () => {
    render(<Inbox />);

    expect(screen.getByText('inbox.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('inbox.emptyDescription')).toBeInTheDocument();

    const refresh = screen.getByRole('button', { name: 'inbox.refreshCta' });
    fireEvent.click(refresh);
    expect(notificationsState.refetch).toHaveBeenCalled();
  });

  it('offers a way back to All when a filter empties the list, instead of Refresh', () => {
    notificationsState.data = [makeNotification({ type: 'assigned', read: true })];
    render(<Inbox />);

    fireEvent.click(screen.getByRole('button', { name: 'inbox.filters.unread' }));

    expect(screen.getByText('inbox.emptyTitle')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'inbox.refreshCta' })).not.toBeInTheDocument();

    const allButtons = screen.getAllByRole('button', { name: 'inbox.filters.all' });
    expect(allButtons.length).toBeGreaterThan(0);
  });
});
