import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Invoices from './Invoices';

const invoice = {
  id: 'INV-100',
  client: 'Acme Corp',
  amount: 1250,
  currency: 'USD',
  paid: 0,
  status: 'Sent' as const,
  dueDate: '2026-08-31',
  issuedDate: '2026-08-01',
};

const { navigateMock, useInvoicesMock } = vi.hoisted(() => ({
  navigateMock: vi.fn(),
  useInvoicesMock: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children?: React.ReactNode }) => <a href="#">{children}</a>,
  useNavigate: () => navigateMock,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('./useInvoices', () => ({
  useInvoices: useInvoicesMock,
  useUpdateInvoiceStatus: () => ({ mutate: vi.fn(), isPending: false }),
  useCreateInvoice: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useUnbilledBillableEntries', () => ({
  useUnbilledBillableEntries: vi.fn(() => ({ data: null })),
  useMarkEntriesInvoiced: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount: number }) => <span>{amount}</span>,
}));

vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/InvoiceStatusFilter', () => ({
  default: () => null,
}));
vi.mock('../../shared/components/DownloadPdfButton', () => ({
  default: () => <button type="button">PDF</button>,
}));
vi.mock('../../shared/components/SendInvoiceModal', () => ({ default: () => null }));
vi.mock('../../shared/lib/exports', () => ({
  exportQuickbooksInvoices: vi.fn(),
  exportXeroInvoices: vi.fn(),
}));

describe('Invoices table layout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useInvoicesMock.mockReturnValue({
      data: { invoices: [invoice], hasMore: false },
      isLoading: false,
    });
  });

  it('keeps the wide table horizontally scrollable for RTL layouts', () => {
    render(<Invoices />);

    const table = screen.getByRole('table');
    const scrollContainer = table.parentElement;

    expect(scrollContainer).not.toBeNull();
    expect(scrollContainer).toHaveStyle({ overflowX: 'auto' });
    expect(table).toHaveStyle({ minWidth: '720px' });
  });
});

describe('Invoices empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useInvoicesMock.mockReturnValue({
      data: { invoices: [], hasMore: false },
      isLoading: false,
    });
  });

  it('renders the shared EmptyState CTA and navigates to the new invoice screen on click', () => {
    render(<Invoices />);

    expect(screen.queryByRole('table')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'invoices.emptyCta' }));

    expect(navigateMock).toHaveBeenCalledWith({ to: '/invoices/new' });
  });
});
