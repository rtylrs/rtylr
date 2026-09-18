import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Reports from './Reports';

initTestI18n();

interface MockUserState {
  user: { role: string | null } | null;
  role: string | null;
}

const mocks = vi.hoisted(() => ({
  role: null as string | null,
  useReportTemplates: vi.fn(),
  useGenerateRevenueReport: vi.fn(),
  useExportRevenueReport: vi.fn(),
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

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (state: MockUserState) => unknown) =>
    selector({ user: { role: mocks.role }, role: mocks.role }),
}));

vi.mock('../../shared/hooks/queries/useInsightsReports', () => ({
  useReportTemplates: mocks.useReportTemplates,
  useGenerateRevenueReport: mocks.useGenerateRevenueReport,
  useExportRevenueReport: mocks.useExportRevenueReport,
}));

describe('Reports export gating', () => {
  beforeEach(() => {
    mocks.role = null;
    mocks.useReportTemplates.mockReturnValue({
      data: { templates: [] },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useGenerateRevenueReport.mockReturnValue({ mutate: vi.fn(), isPending: false });
    mocks.useExportRevenueReport.mockReturnValue({ mutate: vi.fn(), isPending: false });
  });

  it('hides the export button for a non-admin-tier role', () => {
    mocks.role = 'operator';
    render(<Reports />);

    expect(screen.queryByRole('button', { name: 'Export Revenue CSV' })).not.toBeInTheDocument();
    // The generate action is not export-gated and must remain visible.
    expect(screen.getByRole('button', { name: 'Generate Revenue Report' })).toBeInTheDocument();
  });

  it('hides the export button when there is no role at all', () => {
    mocks.role = null;
    render(<Reports />);

    expect(screen.queryByRole('button', { name: 'Export Revenue CSV' })).not.toBeInTheDocument();
  });

  it('shows the export button for admin', () => {
    mocks.role = 'admin';
    render(<Reports />);

    expect(screen.getByRole('button', { name: 'Export Revenue CSV' })).toBeInTheDocument();
  });

  it('shows the export button for owner', () => {
    mocks.role = 'owner';
    render(<Reports />);

    expect(screen.getByRole('button', { name: 'Export Revenue CSV' })).toBeInTheDocument();
  });
});

describe('Reports empty state', () => {
  beforeEach(() => {
    mocks.role = null;
    mocks.useGenerateRevenueReport.mockReturnValue({ mutate: vi.fn(), isPending: false });
    mocks.useExportRevenueReport.mockReturnValue({ mutate: vi.fn(), isPending: false });
  });

  it('shows the illustration, real title/description and a refresh action when there are no templates', () => {
    const refetch = vi.fn();
    mocks.useReportTemplates.mockReturnValue({
      data: { templates: [] },
      isLoading: false,
      isError: false,
      refetch,
    });

    render(<Reports />);

    expect(screen.getByText('No report templates yet')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Report templates will appear here once they are made available for this workspace.'
      )
    ).toBeInTheDocument();

    const refreshButton = screen.getByRole('button', { name: 'Refresh' });
    fireEvent.click(refreshButton);

    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when templates exist', () => {
    mocks.useReportTemplates.mockReturnValue({
      data: { templates: [{ id: '1', name: 'Weekly Revenue', description: '' }] },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });

    render(<Reports />);

    expect(screen.queryByText('No report templates yet')).not.toBeInTheDocument();
  });
});
