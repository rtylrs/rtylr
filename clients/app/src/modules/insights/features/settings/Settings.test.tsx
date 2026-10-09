import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Settings from './Settings';
import { useUserStore } from '../../../../shared/stores/userStore';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

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

vi.mock('../../shared/hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
  useUpdateInsightsSettings: mocks.useUpdateInsightsSettings,
  useScheduledReports: mocks.useScheduledReports,
  useDeleteScheduledReport: mocks.useDeleteScheduledReport,
}));

beforeEach(() => {
  useUserStore.setState({
    user: { effectiveAccess: createEffectiveAccess('insights', { isOwner: true }) },
  });
});
afterEach(() => useUserStore.setState({ user: null }));

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

  it('lets dashboard readers view settings without exposing save or report mutations', () => {
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('insights', {
          permissions: ['insights:dashboard:view'],
        }),
      },
    });
    render(<Settings />);

    expect(screen.getByLabelText('Default Currency')).toHaveAttribute('readonly');
    expect(screen.queryByRole('button', { name: 'Save settings' })).not.toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: 'Scheduled Reports' })).not.toBeInTheDocument();
    expect(mocks.useScheduledReports).toHaveBeenCalledWith({ enabled: false });
    expect(mocks.useUpdateInsightsSettings().mutate).not.toHaveBeenCalled();
  });

  it('allows an assigned settings editor to save without granting report mutations', () => {
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('insights', {
          permissions: ['insights:dashboard:view', 'insights:settings:edit'],
        }),
      },
    });
    render(<Settings />);
    fireEvent.change(screen.getByLabelText('Default Currency'), { target: { value: 'EUR' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }));

    expect(mocks.useUpdateInsightsSettings().mutate).toHaveBeenCalledWith(
      {
        defaultCurrency: 'EUR',
        fiscalYearStart: '01-01',
        timezone: 'UTC',
      },
      { onSuccess: expect.any(Function) }
    );
    expect(mocks.useScheduledReports).toHaveBeenCalledWith({ enabled: false });
  });

  it.each([{ isOwner: true }, { isAdmin: true }])(
    'unlocks settings and report actions for %j',
    (flags) => {
      useUserStore.setState({
        user: { effectiveAccess: createEffectiveAccess('insights', flags) },
      });
      render(<Settings />);

      expect(screen.getByRole('button', { name: 'Save settings' })).toBeEnabled();
      expect(screen.getByLabelText('Default Currency')).not.toHaveAttribute('readonly');
      expect(mocks.useScheduledReports).toHaveBeenCalledWith({ enabled: true });
    }
  );
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

  it('keeps edits while a save is pending', () => {
    const mutate = vi.fn();
    mocks.useUpdateInsightsSettings.mockReturnValue({ mutate, isPending: false });
    settings('USD', '01-01', 'UTC');

    const { rerender } = render(<Settings />);

    fireEvent.change(screen.getByLabelText('Default Currency'), { target: { value: 'EUR' } });
    fireEvent.change(screen.getByLabelText('Timezone'), { target: { value: 'Africa/Cairo' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }));
    expect(mutate).toHaveBeenCalledWith(
      {
        defaultCurrency: 'EUR',
        fiscalYearStart: '01-01',
        timezone: 'Africa/Cairo',
      },
      { onSuccess: expect.any(Function) }
    );

    // A background refetch must not discard unsaved edits.
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
