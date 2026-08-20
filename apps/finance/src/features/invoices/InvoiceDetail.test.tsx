import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import InvoiceDetail from './InvoiceDetail';

const invoice = {
  id: 'INV-100',
  client: 'Acme Corp',
  clientEmail: 'billing@acme.com',
  amount: 1250,
  currency: 'USD',
  paid: 0,
  status: 'Sent' as const,
  dueDate: '2026-08-31',
  issuedDate: '2026-08-01',
  items: [],
  history: [],
};

const mocks = vi.hoisted(() => ({ role: 'staff' as string | null }));

vi.mock('@voxire/auth', () => ({
  useUserStore: (
    selector: (state: { user: { role: string | null }; role: string | null }) => unknown
  ) => selector({ user: { role: mocks.role }, role: mocks.role }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('@voxire/api', () => ({
  getOneKitErrorMessage: (_e: unknown, fallback: string) => fallback,
}));

vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useParams: () => ({ id: 'INV-100' }) }),
  useNavigate: () => vi.fn(),
}));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('./useInvoices', () => ({
  useInvoiceDetail: () => ({ data: invoice, isLoading: false }),
  useDeleteInvoice: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));

vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount: number }) => <span>{amount}</span>,
}));
vi.mock('../../shared/components/PaymentBadge', () => ({ default: () => null }));
vi.mock('../../shared/components/DownloadPdfButton', () => ({
  default: () => <button type="button">PDF</button>,
}));
vi.mock('../../shared/components/InvoicePreview', () => ({ default: () => null }));
vi.mock('../../shared/components/SendInvoiceModal', () => ({ default: () => null }));
vi.mock('../../shared/components/RecordPaymentModal', () => ({ default: () => null }));
vi.mock('../../shared/components/ConfirmDialog', () => ({ default: () => null }));
vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));

describe('InvoiceDetail send/void actions', () => {
  beforeEach(() => {
    mocks.role = 'staff';
  });

  it('hides Send/Void for a role without finance:invoices:send', () => {
    mocks.role = 'staff';
    render(<InvoiceDetail />);

    expect(screen.queryByText('Send')).not.toBeInTheDocument();
    expect(screen.queryByText('Void')).not.toBeInTheDocument();
  });

  it('shows Send/Void for admin', () => {
    mocks.role = 'admin';
    render(<InvoiceDetail />);

    expect(screen.getByText('Send')).toBeInTheDocument();
    expect(screen.getByText('Void')).toBeInTheDocument();
  });

  it('shows Send/Void for owner', () => {
    mocks.role = 'owner';
    render(<InvoiceDetail />);

    expect(screen.getByText('Send')).toBeInTheDocument();
    expect(screen.getByText('Void')).toBeInTheDocument();
  });
});
