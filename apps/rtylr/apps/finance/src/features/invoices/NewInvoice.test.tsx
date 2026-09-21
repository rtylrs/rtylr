import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import NewInvoice from './NewInvoice';
import { initTestI18n } from '../../shared/testing/i18n';

const { createInvoiceMock, listCrmCustomersMock } = vi.hoisted(() => ({
  createInvoiceMock: vi.fn(),
  listCrmCustomersMock: vi.fn(),
}));

vi.mock('../../shared/lib/clients', () => ({
  listCrmCustomers: listCrmCustomersMock,
}));

// The route's search params are the input under test: the same shape the
// global search links to (/invoices/new?client=<id>).
vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => ({ clone: undefined, client: '42' }) }),
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
  useCreateInvoice: () => ({ mutate: createInvoiceMock, isPending: false }),
  useUpdateInvoiceStatus: () => ({ mutate: vi.fn(), isPending: false }),
  useInvoiceDetail: () => ({ data: undefined, isLoading: false }),
}));

vi.mock('../../shared/hooks/queries/useSignedContracts', () => ({
  useSignedContracts: () => ({ data: null }),
}));

vi.mock('../../shared/components/PageHeader', () => ({ default: () => null }));

// i18next is not mocked here: the real locale files mean an assertion
// written in English also proves the key resolves (rtylr#1269).
beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  createInvoiceMock.mockReset();
  listCrmCustomersMock.mockReset();
});

describe('NewInvoice deep-linked client', () => {
  it('carries the CRM customer id from a ?client= deep link into the create payload', async () => {
    listCrmCustomersMock.mockResolvedValue({
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

    expect(createInvoiceMock).toHaveBeenCalledTimes(1);
    expect(createInvoiceMock.mock.calls[0][0]).toMatchObject({
      clientName: 'Acme Corp',
      customerId: '42',
    });
  });
});
