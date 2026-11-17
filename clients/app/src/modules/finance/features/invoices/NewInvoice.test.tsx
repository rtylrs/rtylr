import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import NewInvoice from './NewInvoice';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  createInvoiceMock: vi.fn(),
  listCrmCustomersMock: vi.fn(),
  routeSearch: { clone: undefined as string | undefined, client: '42' as string | undefined },
  sourceInvoice: undefined as unknown,
  workProjects: null as { id: string; name: string }[] | null,
}));

vi.mock('../../shared/lib/clients', () => ({
  listCrmCustomers: mocks.listCrmCustomersMock,
}));

// The route's search params are the input under test: the same shape the
// global search links to (/invoices/new?client=<id>).
vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => mocks.routeSearch }),
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => Promise.resolve(value),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('./useInvoices', () => ({
  useCreateInvoice: () => ({ mutate: mocks.createInvoiceMock, isPending: false }),
  useUpdateInvoiceStatus: () => ({ mutate: vi.fn(), isPending: false }),
  useInvoiceDetail: () => ({ data: mocks.sourceInvoice, isLoading: false }),
}));

vi.mock('../../shared/hooks/queries/useSignedContracts', () => ({
  useSignedContracts: () => ({ data: null }),
}));

vi.mock('../../shared/hooks/queries/useWorkProjects', () => ({
  useWorkProjects: () => ({ data: mocks.workProjects }),
}));

vi.mock('../../shared/components/PageHeader', () => ({ default: () => null }));

// i18next is not mocked here: the real locale files mean an assertion
// written in English also proves the key resolves (rtylr#1269).
beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  mocks.createInvoiceMock.mockReset();
  mocks.listCrmCustomersMock.mockReset();
  mocks.routeSearch = { clone: undefined, client: '42' };
  mocks.sourceInvoice = undefined;
  mocks.workProjects = null;
});

// Found by id rather than label so the picker tests do not depend on the
// locale strings, which land in their own commits.
const projectPicker = () => document.getElementById('invoice-project') as HTMLSelectElement;

const fillLineAndSave = () => {
  fireEvent.click(screen.getByRole('button', { name: /Next/ }));
  fireEvent.change(screen.getByPlaceholderText('Description'), {
    target: { value: 'Consulting' },
  });
  const [, priceInput] = screen.getAllByRole('textbox');
  fireEvent.change(priceInput, { target: { value: '100' } });
  fireEvent.click(screen.getByRole('button', { name: /Next/ }));
  fireEvent.click(screen.getByRole('button', { name: /Next/ }));
  fireEvent.click(screen.getByRole('button', { name: /Save as Draft/ }));
};

describe('NewInvoice project picker', () => {
  it('sends the chosen Work project with the create payload', async () => {
    mocks.listCrmCustomersMock.mockResolvedValue({
      customers: [{ id: 42, name: 'Acme Corp', email: 'billing@acme.test' }],
    });
    mocks.workProjects = [
      { id: '7', name: 'Website' },
      { id: '8', name: 'App redesign' },
    ];

    render(<NewInvoice />);
    await screen.findByText('Acme Corp');

    fireEvent.change(projectPicker(), { target: { value: '8' } });
    fillLineAndSave();

    expect(mocks.createInvoiceMock.mock.calls[0]?.[0]).toMatchObject({ projectId: '8' });
  });

  it('leaves the invoice untagged by default', async () => {
    mocks.listCrmCustomersMock.mockResolvedValue({
      customers: [{ id: 42, name: 'Acme Corp', email: 'billing@acme.test' }],
    });
    mocks.workProjects = [{ id: '7', name: 'Website' }];

    render(<NewInvoice />);
    await screen.findByText('Acme Corp');
    fillLineAndSave();

    expect(mocks.createInvoiceMock.mock.calls[0]?.[0].projectId).toBeUndefined();
  });

  it('disables the picker when no Work projects are available', async () => {
    mocks.listCrmCustomersMock.mockResolvedValue({ customers: [] });

    render(<NewInvoice />);

    const picker = projectPicker();
    expect(picker).toBeDisabled();
    expect(picker.options).toHaveLength(1);
    expect(picker.value).toBe('');
  });

  it('keeps the cloned invoice project, even one no longer listed', async () => {
    mocks.routeSearch = { clone: 'INV-9', client: undefined };
    mocks.workProjects = [{ id: '7', name: 'Website' }];
    mocks.sourceInvoice = {
      id: 'INV-9',
      client: 'Source Client',
      clientEmail: 'source@example.test',
      amount: 100,
      status: 'Sent',
      issuedDate: '2026-08-01',
      dueDate: '2026-08-31',
      projectId: '99',
      items: [{ id: 'source-1', description: 'Line', quantity: 1, unitPrice: 100, taxRate: 0 }],
    };

    render(<NewInvoice />);
    await screen.findByText('Source Client');
    expect(projectPicker()).toHaveValue('99');

    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Save as Draft/ }));

    expect(mocks.createInvoiceMock.mock.calls[0]?.[0]).toMatchObject({ projectId: '99' });
  });
});

describe('NewInvoice deep-linked client', () => {
  it('carries the CRM customer id from a ?client= deep link into the create payload', async () => {
    mocks.listCrmCustomersMock.mockResolvedValue({
      customers: [{ id: 42, name: 'Acme Corp', email: 'billing@acme.test' }],
    });

    render(<NewInvoice />);

    // The deep link resolves against the CRM directory and preselects the client.
    await screen.findByText('Acme Corp');

    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.change(screen.getByPlaceholderText('Description'), {
      target: { value: 'Consulting' },
    });
    const [, priceInput] = screen.getAllByRole('textbox');
    fireEvent.change(priceInput, { target: { value: '100' } });
    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Save as Draft/ }));

    expect(mocks.createInvoiceMock).toHaveBeenCalledTimes(1);
    expect(mocks.createInvoiceMock.mock.calls[0][0]).toMatchObject({
      clientName: 'Acme Corp',
      customerId: '42',
    });
  });

  it('preserves the cloned line tax rate and percentage discount', async () => {
    mocks.routeSearch = { clone: 'INV-9', client: undefined };
    mocks.sourceInvoice = {
      id: 'INV-9',
      client: 'Source Client',
      clientEmail: 'source@example.test',
      amount: 90,
      currency: 'USD',
      paid: 0,
      status: 'Sent',
      issuedDate: '2026-08-01',
      dueDate: '2026-08-31',
      notes: '',
      discount: 10,
      discountType: 'percentage',
      items: [
        {
          id: 'source-1',
          description: 'Taxed line',
          quantity: 1,
          unitPrice: 100,
          taxRate: 17.5,
        },
      ],
    };

    render(<NewInvoice />);
    await screen.findByText('Source Client');

    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Next/ }));
    fireEvent.click(screen.getByRole('button', { name: /Save as Draft/ }));

    expect(mocks.createInvoiceMock.mock.calls[0]?.[0]).toEqual(
      expect.objectContaining({
        discountType: 'percentage',
        discountValue: 10,
        items: [expect.objectContaining({ taxRate: 17.5 })],
      })
    );
  });
});
