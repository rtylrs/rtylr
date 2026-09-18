import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => (opts?.name ? `${key}:${opts.name}` : key),
  }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

const setManagerMutate = vi.hoisted(() => vi.fn());
const record = vi.hoisted(() => ({ value: { id: '1', firstName: 'Ada', lastName: 'L' } }));

vi.mock('./useStaff', () => ({
  useStaffReportingLines: () => ({
    data: [
      { id: 1, name: 'Ada', managerStaffId: 0 },
      { id: 2, name: 'Grace', managerStaffId: 1 },
    ],
    isLoading: false,
  }),
  useStaffRecord: () => ({ data: record.value }),
  useSetStaffManager: () => ({ mutate: setManagerMutate, isPending: false }),
}));

import ReportingLines from './ReportingLines';

beforeEach(() => setManagerMutate.mockClear());

describe('ReportingLines', () => {
  it('shows who each person reports to, and "no manager" when unset', () => {
    render(<ReportingLines />);

    expect(
      screen.getByText('staff.reportingLines.reportsTo:staff.reportingLines.none')
    ).toBeInTheDocument();
    expect(screen.getByText('staff.reportingLines.reportsTo:Ada')).toBeInTheDocument();
  });

  // Setting a manager widens what they can see in Flow, so the screen has to say
  // so rather than leaving it as a surprise from an HR form.
  it('states the Flow visibility consequence', () => {
    render(<ReportingLines />);

    expect(screen.getByText('staff.reportingLines.flowNotice')).toBeInTheDocument();
  });

  it('sends the full record with the chosen manager', () => {
    render(<ReportingLines />);
    fireEvent.click(screen.getByLabelText('staff.reportingLines.changeLabel:Ada'));
    fireEvent.change(screen.getByLabelText('staff.reportingLines.selectLabel:Ada'), {
      target: { value: '2' },
    });

    expect(setManagerMutate).toHaveBeenCalledWith(
      { staff: record.value, managerStaffId: 2 },
      expect.anything()
    );
  });

  // Nobody reports to themselves, and the resolver excludes it too, so the UI
  // must not offer an option the server would ignore.
  it('leaves the person out of their own manager list', () => {
    render(<ReportingLines />);
    fireEvent.click(screen.getByLabelText('staff.reportingLines.changeLabel:Ada'));

    const select = screen.getByLabelText('staff.reportingLines.selectLabel:Ada');
    const values = Array.from(select.querySelectorAll('option')).map((o) =>
      o.getAttribute('value')
    );
    expect(values).not.toContain('1');
    expect(values).toContain('2');
  });
});
