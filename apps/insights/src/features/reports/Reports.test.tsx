import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import i18next from 'i18next';
import { initReactI18next } from 'react-i18next';
import enInsights from '../../locales/en/insights.json';
import Reports from './Reports';

// This app's i18n bootstrap (`shared/services/i18next.ts`) merges every
// locales/en/*.json file into a single 'translation' namespace keyed by
// filename, so `t('insights.reports.exportRevenue')` resolves as a dot-path
// within that namespace. Reproduce that here with the real, shipped English
// copy so assertions check actual production strings, not raw i18n keys.
void i18next.use(initReactI18next).init({
  lng: 'en',
  fallbackLng: 'en',
  resources: { en: { translation: { insights: enInsights } } },
  interpolation: { escapeValue: false },
});

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
