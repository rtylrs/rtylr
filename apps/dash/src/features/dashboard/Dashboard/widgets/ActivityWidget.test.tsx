import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

const activityData = vi.hoisted(() => ({
  data: [] as { id: string }[],
  refetch: vi.fn(),
}));
vi.mock('../../../../shared/hooks/queries/useActivity', () => ({
  useActivity: () => activityData,
}));

vi.mock('../../../../shared/components/dash', () => ({
  ActivityEntryRow: ({ entry }: { entry: { id: string } }) => <div>{entry.id}</div>,
}));

import ActivityWidget from './ActivityWidget';

beforeEach(() => {
  vi.clearAllMocks();
  activityData.data = [];
});

describe('ActivityWidget empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there is no activity', () => {
    render(<ActivityWidget />);

    expect(screen.getByText('dashboard.activityWidget.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('dashboard.activityWidget.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', () => {
    render(<ActivityWidget />);

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(activityData.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when there is activity', () => {
    activityData.data = [{ id: '1' }];
    render(<ActivityWidget />);

    expect(screen.queryByText('dashboard.activityWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
