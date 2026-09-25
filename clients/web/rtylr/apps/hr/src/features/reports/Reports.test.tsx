import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReportPayload } from '../../shared/schemas/reports';

const mocks = vi.hoisted(() => ({
  useGenerateReport: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('./useReports', () => ({
  useGenerateReport: mocks.useGenerateReport,
}));

// The label under test is the date range Reports hands to the table; the
// table's own rendering is covered by ReportTable.test.tsx.
vi.mock('./ReportTable', () => ({
  default: ({ dateFrom, dateTo }: { dateFrom?: string; dateTo?: string }) => (
    <div data-testid="report-table" data-from={dateFrom} data-to={dateTo} />
  ),
}));

vi.mock('motion/react', () => ({
  motion: new Proxy(
    {},
    {
      get:
        () =>
        ({ children }: React.PropsWithChildren) => <div>{children}</div>,
    }
  ),
}));

import Reports from './Reports';

const payload: ReportPayload = {
  reportType: 'attendance_summary',
  columns: [],
  rows: [],
  generatedAt: 1700000000,
};

describe('Reports generated range', () => {
  const mutate = vi.fn();

  beforeEach(() => {
    mutate.mockReset();
    mutate.mockImplementation(
      (_vars: unknown, opts: { onSuccess: (data: ReportPayload) => void }) => {
        opts.onSuccess(payload);
      }
    );
    mocks.useGenerateReport.mockReturnValue({ mutate, isPending: false });
  });

  it('keeps the generated range when the filter dates change after Generate', () => {
    render(<Reports />);

    fireEvent.change(screen.getByLabelText('reports.dateFrom'), {
      target: { value: '2026-01-01' },
    });
    fireEvent.change(screen.getByLabelText('reports.dateTo'), { target: { value: '2026-01-31' } });
    fireEvent.click(screen.getByRole('button', { name: 'reports.generate' }));

    expect(screen.getByTestId('report-table')).toHaveAttribute('data-from', '2026-01-01');
    expect(screen.getByTestId('report-table')).toHaveAttribute('data-to', '2026-01-31');

    // Editing the live filter must not relabel a report that was already
    // generated: the exported PDF prints these values as its period.
    fireEvent.change(screen.getByLabelText('reports.dateTo'), { target: { value: '2026-02-28' } });

    expect(screen.getByTestId('report-table')).toHaveAttribute('data-from', '2026-01-01');
    expect(screen.getByTestId('report-table')).toHaveAttribute('data-to', '2026-01-31');
  });

  it('clears the previous result when another report type is selected', () => {
    render(<Reports />);

    fireEvent.click(screen.getByRole('button', { name: 'reports.generate' }));
    expect(screen.getByTestId('report-table')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'reports.reportTypes.leave_balance' }));

    expect(screen.queryByTestId('report-table')).not.toBeInTheDocument();
  });
});
