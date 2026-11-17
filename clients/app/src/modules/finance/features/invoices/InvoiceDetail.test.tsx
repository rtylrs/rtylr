import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import InvoiceDetail from './InvoiceDetail';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

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

const mocks = vi.hoisted(() => ({
  effectiveAccess: undefined as unknown,
  projectId: undefined as string | undefined,
  workProjects: null as { id: string; name: string }[] | null,
  status: 'Sent' as 'Draft' | 'Sent' | 'Paid' | 'Overdue' | 'Cancelled',
  setProject: vi.fn(),
}));

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: Object.assign(
    (
      selector: (state: { user: { effectiveAccess: unknown }; effectiveAccess: unknown }) => unknown
    ) =>
      selector({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    {
      getState: () => ({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    }
  ),
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
  useInvoiceDetail: () => ({
    data: { ...invoice, status: mocks.status, projectId: mocks.projectId },
    isLoading: false,
  }),
  useDeleteInvoice: () => ({ mutate: vi.fn(), isPending: false }),
  useSetInvoiceProject: () => ({ mutate: mocks.setProject, isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useWorkProjects', () => ({
  useWorkProjects: () => ({ data: mocks.workProjects }),
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
    mocks.effectiveAccess = createEffectiveAccess('finance', {
      permissions: ['finance:invoices:view'],
    });
  });

  it('hides Send/Void for a role without finance:invoices:send', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', {
      permissions: ['finance:invoices:view'],
    });
    render(<InvoiceDetail />);

    expect(screen.queryByText('Send')).not.toBeInTheDocument();
    expect(screen.queryByText('Void')).not.toBeInTheDocument();
  });

  it('shows Send/Void for admin', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', { isAdmin: true });
    render(<InvoiceDetail />);

    expect(screen.getByText('Send')).toBeInTheDocument();
    expect(screen.getByText('Void')).toBeInTheDocument();
  });

  it('shows Send/Void for owner', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', { isOwner: true });
    render(<InvoiceDetail />);

    expect(screen.getByText('Send')).toBeInTheDocument();
    expect(screen.getByText('Void')).toBeInTheDocument();
  });
});

describe('InvoiceDetail project', () => {
  beforeEach(() => {
    mocks.effectiveAccess = createEffectiveAccess('finance', {
      permissions: ['finance:invoices:view'],
    });
    mocks.projectId = undefined;
    mocks.workProjects = null;
    mocks.status = 'Sent';
  });

  it('shows the tagged Work project by name', () => {
    mocks.projectId = '7';
    mocks.workProjects = [{ id: '7', name: 'Website' }];
    render(<InvoiceDetail />);

    expect(screen.getByText('Website')).toBeInTheDocument();
  });

  it('falls back to the project id when the project is not listed', () => {
    mocks.projectId = '99';
    render(<InvoiceDetail />);

    expect(screen.getByText('financeInvoices.new.projectFallback')).toBeInTheDocument();
  });

  it('shows no project card for an untagged invoice the viewer cannot change', () => {
    render(<InvoiceDetail />);

    expect(screen.queryByText('financeInvoices.detail.project')).not.toBeInTheDocument();
  });
});

describe('InvoiceDetail change project', () => {
  beforeEach(() => {
    mocks.effectiveAccess = createEffectiveAccess('finance', {
      permissions: ['finance:invoices:view', 'finance:invoices:create'],
    });
    mocks.projectId = undefined;
    mocks.workProjects = [
      { id: '7', name: 'Website' },
      { id: '8', name: 'App redesign' },
    ];
    mocks.status = 'Sent';
    mocks.setProject.mockReset();
  });

  it('tags a sent, untagged invoice with a Work project', () => {
    render(<InvoiceDetail />);

    expect(screen.getByText('financeInvoices.new.noProject')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /financeInvoices.detail.changeProject/ }));
    const select = screen.getByLabelText('financeInvoices.detail.changeProject');
    expect(
      Array.from((select as HTMLSelectElement).options).map((option) => option.textContent)
    ).toEqual(['financeInvoices.new.noProject', 'Website', 'App redesign']);

    fireEvent.change(select, { target: { value: '8' } });
    expect(mocks.setProject).toHaveBeenCalledWith(
      { id: 'INV-100', projectId: '8' },
      expect.objectContaining({ onSuccess: expect.any(Function) })
    );
  });

  it('clears the project with No project', () => {
    mocks.projectId = '7';
    render(<InvoiceDetail />);

    fireEvent.click(screen.getByRole('button', { name: /financeInvoices.detail.changeProject/ }));
    fireEvent.change(screen.getByLabelText('financeInvoices.detail.changeProject'), {
      target: { value: '' },
    });
    expect(mocks.setProject).toHaveBeenCalledWith(
      { id: 'INV-100', projectId: null },
      expect.anything()
    );
  });

  it('does not write when the same project is picked again', () => {
    mocks.projectId = '7';
    render(<InvoiceDetail />);

    fireEvent.click(screen.getByRole('button', { name: /financeInvoices.detail.changeProject/ }));
    fireEvent.change(screen.getByLabelText('financeInvoices.detail.changeProject'), {
      target: { value: '7' },
    });
    expect(mocks.setProject).not.toHaveBeenCalled();
    expect(screen.queryByLabelText('financeInvoices.detail.changeProject')).not.toBeInTheDocument();
  });

  it('keeps an unlisted current project selectable', () => {
    mocks.projectId = '99';
    render(<InvoiceDetail />);

    fireEvent.click(screen.getByRole('button', { name: /financeInvoices.detail.changeProject/ }));
    const select = screen.getByLabelText(
      'financeInvoices.detail.changeProject'
    ) as HTMLSelectElement;
    expect(select.value).toBe('99');
  });

  it('offers the change on a paid invoice', () => {
    mocks.status = 'Paid';
    render(<InvoiceDetail />);

    expect(
      screen.getByRole('button', { name: /financeInvoices.detail.changeProject/ })
    ).toBeInTheDocument();
  });

  it('hides the change on a cancelled or void invoice', () => {
    mocks.status = 'Cancelled';
    mocks.projectId = '7';
    render(<InvoiceDetail />);

    expect(screen.getByText('Website')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: /financeInvoices.detail.changeProject/ })
    ).not.toBeInTheDocument();
  });

  it('hides the change without finance:invoices:create', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', {
      permissions: ['finance:invoices:view', 'finance:invoices:send'],
    });
    mocks.projectId = '7';
    render(<InvoiceDetail />);

    expect(screen.getByText('Website')).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: /financeInvoices.detail.changeProject/ })
    ).not.toBeInTheDocument();
  });
});
