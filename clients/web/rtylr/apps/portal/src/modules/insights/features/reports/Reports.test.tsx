import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { EffectiveAccess } from '@voxire/types';
import { createEffectiveAccess } from '../../../../../../../test/effectiveAccess';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Reports from './Reports';
import { useUserStore } from '../../../../shared/stores/userStore';

initTestI18n();

const mocks = vi.hoisted(() => ({
  effectiveAccess: null as EffectiveAccess | null,
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

vi.mock('../../shared/hooks/queries/useInsightsReports', () => ({
  useReportTemplates: mocks.useReportTemplates,
  useGenerateRevenueReport: mocks.useGenerateRevenueReport,
  useExportRevenueReport: mocks.useExportRevenueReport,
}));

function renderReports() {
  useUserStore.setState({
    user: mocks.effectiveAccess ? { effectiveAccess: mocks.effectiveAccess } : null,
  });
  return render(<Reports />);
}

describe('Reports export gating', () => {
  beforeEach(() => {
    mocks.effectiveAccess = createEffectiveAccess('insights', {
      permissions: ['insights:reports:create'],
    });
    mocks.useReportTemplates.mockReturnValue({
      data: { templates: [] },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useGenerateRevenueReport.mockReturnValue({ mutate: vi.fn(), isPending: false });
    mocks.useExportRevenueReport.mockReturnValue({ mutate: vi.fn(), isPending: false });
  });

  it('hides the export button when a member has create but not export permission', () => {
    mocks.effectiveAccess = createEffectiveAccess('insights', {
      permissions: ['insights:reports:create'],
    });
    renderReports();

    expect(screen.queryByRole('button', { name: 'Export Revenue CSV' })).not.toBeInTheDocument();
    // The generate action is not export-gated and must remain visible.
    expect(screen.getByRole('button', { name: 'Generate Revenue Report' })).toBeInTheDocument();
  });

  it('hides the export button when authorization data is missing', () => {
    mocks.effectiveAccess = null;
    renderReports();

    expect(screen.queryByRole('button', { name: 'Export Revenue CSV' })).not.toBeInTheDocument();
  });

  it('shows the export button for admin', () => {
    mocks.effectiveAccess = createEffectiveAccess('insights', { isAdmin: true });
    renderReports();

    expect(screen.getByRole('button', { name: 'Export Revenue CSV' })).toBeInTheDocument();
  });

  it('shows the export button for owner', () => {
    mocks.effectiveAccess = createEffectiveAccess('insights', { isOwner: true });
    renderReports();

    expect(screen.getByRole('button', { name: 'Export Revenue CSV' })).toBeInTheDocument();
  });
});

describe('Reports empty state', () => {
  beforeEach(() => {
    mocks.effectiveAccess = createEffectiveAccess('insights', {
      permissions: ['insights:reports:create'],
    });
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

    renderReports();

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

    renderReports();

    expect(screen.queryByText('No report templates yet')).not.toBeInTheDocument();
  });
});
