import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Operations from './Operations';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useOperationsKPIs: vi.fn(),
  useHeadcount: vi.fn(),
  useShiftMetrics: vi.fn(),
  useTimeOff: vi.fn(),
  useUtilization: vi.fn(),
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

vi.mock('../../shared/hooks/queries/useInsightsOperations', () => ({
  useOperationsKPIs: mocks.useOperationsKPIs,
  useHeadcount: mocks.useHeadcount,
  useShiftMetrics: mocks.useShiftMetrics,
  useTimeOff: mocks.useTimeOff,
  useUtilization: mocks.useUtilization,
}));

describe('Operations empty states', () => {
  const refetchHeadcount = vi.fn();
  const refetchShifts = vi.fn();
  const refetchTimeOff = vi.fn();
  const refetchUtilization = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useOperationsKPIs.mockReturnValue({
      data: { activeStaff: 0, shiftsThisWeek: 0, pendingTimeOff: 0, openLeaveRequests: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useHeadcount.mockReturnValue({
      data: undefined,
      isLoading: false,
      refetch: refetchHeadcount,
    });
    mocks.useShiftMetrics.mockReturnValue({
      data: undefined,
      isLoading: false,
      refetch: refetchShifts,
    });
    mocks.useTimeOff.mockReturnValue({
      data: undefined,
      isLoading: false,
      refetch: refetchTimeOff,
    });
    mocks.useUtilization.mockReturnValue({
      data: undefined,
      isLoading: false,
      refetch: refetchUtilization,
    });
  });

  it('shows a distinct title, description and working refresh action for headcount', () => {
    render(<Operations />);

    expect(screen.getByText('No headcount data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Headcount will appear here once staff records are added.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[0]);
    expect(refetchHeadcount).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for utilization', () => {
    render(<Operations />);

    expect(screen.getByText('No utilization data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Utilization will appear here once shifts are scheduled and worked.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[1]);
    expect(refetchUtilization).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for shift metrics', () => {
    render(<Operations />);

    expect(screen.getByText('No shift data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Shift metrics will appear here once shifts are scheduled.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[2]);
    expect(refetchShifts).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for time off', () => {
    render(<Operations />);

    expect(screen.getByText('No time-off data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Time-off requests will appear here once staff submit them.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[3]);
    expect(refetchTimeOff).toHaveBeenCalledTimes(1);
  });
});
