import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Settings from './Settings';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useInsightsSettings: vi.fn(),
  useUpdateInsightsSettings: vi.fn(),
  useScheduledReports: vi.fn(),
  useDeleteScheduledReport: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => (
      <div {...props}>{children}</div>
    ),
  },
}));

vi.mock('../../shared/hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
  useUpdateInsightsSettings: mocks.useUpdateInsightsSettings,
  useScheduledReports: mocks.useScheduledReports,
  useDeleteScheduledReport: mocks.useDeleteScheduledReport,
}));

describe('Settings empty state', () => {
  const refetchScheduled = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useInsightsSettings.mockReturnValue({
      data: { defaultCurrency: 'USD', fiscalYearStart: '01-01', timezone: 'UTC' },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useUpdateInsightsSettings.mockReturnValue({ mutate: vi.fn(), isPending: false });
    mocks.useScheduledReports.mockReturnValue({
      data: { scheduledReports: [] },
      isLoading: false,
      refetch: refetchScheduled,
    });
    mocks.useDeleteScheduledReport.mockReturnValue({ mutate: vi.fn(), isPending: false });
  });

  it('shows the illustration, real title/description and a refresh action when there are no scheduled reports', () => {
    render(<Settings />);

    expect(screen.getByText('No scheduled reports yet')).toBeInTheDocument();
    expect(
      screen.getByText('Reports you schedule for recurring delivery will appear here.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(refetchScheduled).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when scheduled reports exist', () => {
    mocks.useScheduledReports.mockReturnValue({
      data: {
        scheduledReports: [
          {
            id: '1',
            templateId: 'weekly-revenue',
            cadence: 'weekly',
            channel: 'email',
            recipient: 'owner@example.com',
            isActive: true,
            createdAt: '2026-01-01',
          },
        ],
      },
      isLoading: false,
      refetch: refetchScheduled,
    });

    render(<Settings />);

    expect(screen.queryByText('No scheduled reports yet')).not.toBeInTheDocument();
  });
});
