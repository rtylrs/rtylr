import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReportPayload } from '../../shared/schemas/reports';

const mocks = vi.hoisted(() => ({
  useHRConfigStore: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('../../shared/stores/hrConfigStore', () => ({
  useHRConfigStore: mocks.useHRConfigStore,
}));

// happy-dom reports a zero client height for the scroll container, so the
// real virtualizer computes an empty visible range regardless of row count.
// Stub it to render every row - this test cares about the empty-state
// wiring, not virtualization itself.
vi.mock('@tanstack/react-virtual', () => ({
  useVirtualizer: (opts: { count: number; estimateSize: () => number }) => ({
    getTotalSize: () => opts.count * opts.estimateSize(),
    getVirtualItems: () =>
      Array.from({ length: opts.count }, (_, index) => ({
        index,
        start: index * opts.estimateSize(),
        size: opts.estimateSize(),
      })),
  }),
}));

import ReportTable from './ReportTable';

const emptyReport: ReportPayload = {
  reportType: 'headcount',
  columns: [],
  rows: [],
  generatedAt: 1700000000,
};

const filledReport: ReportPayload = {
  reportType: 'headcount',
  columns: ['Name', 'Department'],
  rows: [{ Name: 'Alex Chen', Department: 'FOH' }],
  generatedAt: 1700000000,
};

describe('ReportTable empty state', () => {
  beforeEach(() => {
    mocks.useHRConfigStore.mockReset();
    mocks.useHRConfigStore.mockImplementation((selector: (state: { config: unknown }) => unknown) =>
      selector({ config: {} })
    );
  });

  it('shows the empty state and calls onRefresh when the report has no rows', () => {
    const onRefresh = vi.fn();

    render(<ReportTable report={emptyReport} onRefresh={onRefresh} />);

    expect(screen.getByText('reports.noResults')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(onRefresh).toHaveBeenCalledOnce();
  });

  it('renders the table instead of the empty state when the report has rows', () => {
    const onRefresh = vi.fn();

    render(<ReportTable report={filledReport} onRefresh={onRefresh} />);

    expect(screen.queryByText('reports.noResults')).not.toBeInTheDocument();
    expect(screen.getByText('Alex Chen')).toBeInTheDocument();
    expect(screen.getByText('Name')).toBeInTheDocument();
  });
});
