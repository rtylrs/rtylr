// @vitest-environment happy-dom
import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import RecurringInvoices from './RecurringInvoices';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  role: 'admin' as string | null,
  useRecurringInvoices: vi.fn(),
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (
    selector: (state: { user: { role: string | null }; role: string | null }) => unknown
  ) => selector({ user: { role: mocks.role }, role: mocks.role }),
}));

vi.mock('./useRecurringInvoices', () => ({
  useRecurringInvoices: mocks.useRecurringInvoices,
  useSetRecurringStatus: () => ({ mutate: vi.fn(), isPending: false }),
  useCreateRecurringInvoice: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../invoices/useInvoices', () => ({
  useInvoices: () => ({ data: { invoices: [], hasMore: false }, isFetching: false }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));
vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));
vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/Modal', () => ({ default: () => null }));
vi.mock('@voxire/ui', () => ({
  EmptyState: ({ action }: { action?: { label: string } }) => <div>{action?.label}</div>,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  vi.clearAllMocks();
  mocks.role = 'admin';
  mocks.useRecurringInvoices.mockReturnValue({
    data: {
      invoices: [
        {
          id: '1',
          templateInvoiceId: '42',
          frequency: 'monthly',
          nextInvoiceDate: '2026-07-01',
          isActive: true,
        },
      ],
      hasMore: true,
      page: 1,
      pageSize: 50,
    },
    isLoading: false,
    isError: false,
    refetch: vi.fn(),
  });
});

describe('RecurringInvoices', () => {
  it('loads the next page when the server reports more schedules', () => {
    render(<RecurringInvoices />);

    fireEvent.click(screen.getByRole('button', { name: 'Next' }));

    expect(mocks.useRecurringInvoices).toHaveBeenLastCalledWith({ page: 2 });
  });

  it('hides create and pause controls from a viewer', () => {
    mocks.role = 'viewer';

    render(<RecurringInvoices />);

    expect(screen.queryByRole('button', { name: 'New Recurring Invoice' })).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'Pause recurring invoice 1' })
    ).not.toBeInTheDocument();
  });
});
