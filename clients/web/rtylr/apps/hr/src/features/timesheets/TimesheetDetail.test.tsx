import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  useGetTimesheet: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '5' }),
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../shared/hooks/queries/useTimesheets', () => ({
  useGetTimesheet: mocks.useGetTimesheet,
  useSubmitTimesheet: () => ({ mutate: vi.fn(), isPending: false }),
}));

import TimesheetDetail from './TimesheetDetail';

describe('TimesheetDetail rejection card', () => {
  beforeEach(() => {
    mocks.useGetTimesheet.mockReset();
  });

  it('does not claim the approver left no reason', () => {
    mocks.useGetTimesheet.mockReturnValue({
      data: {
        id: '5',
        status: 'rejected',
        periodStart: '2026-01-05',
        periodEnd: '2026-01-11',
        totalMinutes: 2400,
        entries: [],
      },
      isLoading: false,
    });

    render(<TimesheetDetail />);

    expect(screen.queryByText('timesheets.noReasonProvided')).not.toBeInTheDocument();
    expect(screen.getByText('timesheets.rejectionReasonUnavailable')).toBeInTheDocument();
  });

  it('shows no rejection card for a timesheet that was not rejected', () => {
    mocks.useGetTimesheet.mockReturnValue({
      data: {
        id: '5',
        status: 'approved',
        periodStart: '2026-01-05',
        periodEnd: '2026-01-11',
        totalMinutes: 2400,
        entries: [],
      },
      isLoading: false,
    });

    render(<TimesheetDetail />);

    expect(screen.queryByText('timesheets.rejectionReason')).not.toBeInTheDocument();
  });
});
