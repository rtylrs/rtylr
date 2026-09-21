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

describe('Settings form seeding', () => {
  const settings = (defaultCurrency: string, fiscalYearStart: string, timezone: string): void => {
    mocks.useInsightsSettings.mockReturnValue({
      data: { defaultCurrency, fiscalYearStart, timezone },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
  };

  it('keeps the values the user entered when the save-triggered refetch returns defaults', () => {
    const mutate = vi.fn();
    mocks.useUpdateInsightsSettings.mockReturnValue({ mutate, isPending: false });
    settings('USD', '01-01', 'UTC');

    const { rerender } = render(<Settings />);

    fireEvent.change(screen.getByLabelText('Default Currency'), { target: { value: 'EUR' } });
    fireEvent.change(screen.getByLabelText('Timezone'), { target: { value: 'Africa/Cairo' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }));
    expect(mutate).toHaveBeenCalledWith({
      defaultCurrency: 'EUR',
      fiscalYearStart: '01-01',
      timezone: 'Africa/Cairo',
    });

    // onSuccess invalidates ['insights-settings']; timezone has no server field,
    // so the refetched object carries the default again.
    settings('EUR', '01-01', 'UTC');
    rerender(<Settings />);

    expect(screen.getByLabelText('Default Currency')).toHaveValue('EUR');
    expect(screen.getByLabelText('Timezone')).toHaveValue('Africa/Cairo');
  });

  it('re-seeds untouched fields when the refetched settings differ', () => {
    settings('USD', '01-01', 'UTC');

    const { rerender } = render(<Settings />);
    expect(screen.getByLabelText('Default Currency')).toHaveValue('USD');

    settings('GBP', '04-01', 'Europe/London');
    rerender(<Settings />);

    expect(screen.getByLabelText('Default Currency')).toHaveValue('GBP');
    expect(screen.getByLabelText('Fiscal Year Start')).toHaveValue('04-01');
    expect(screen.getByLabelText('Timezone')).toHaveValue('Europe/London');
  });
});
