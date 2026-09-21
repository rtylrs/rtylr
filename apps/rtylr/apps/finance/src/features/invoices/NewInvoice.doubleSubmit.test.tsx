import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import NewInvoice from './NewInvoice';
import { initTestI18n } from '../../shared/testing/i18n';

/**
 * Save & Send is two mutations (create, then status -> Sent). The invoice
 * create route is not idempotent, so a second submit while the status call is
 * still in flight would persist a duplicate invoice. These tests drive the
 * real useInvoices mutations rather than mocking them, because the bug lived
 * in the pending flags they expose.
 */
const { createInvoiceMock, updateInvoiceStatusMock, listCrmCustomersMock } = vi.hoisted(() => ({
  createInvoiceMock: vi.fn(),
  updateInvoiceStatusMock: vi.fn(),
  listCrmCustomersMock: vi.fn(),
}));

vi.mock('../../shared/lib/clients', () => ({
  listCrmCustomers: listCrmCustomersMock,
}));

vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => ({ clone: undefined, client: '42' }) }),
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    createInvoice: createInvoiceMock,
    updateInvoiceStatus: updateInvoiceStatusMock,
    listInvoices: vi.fn(),
  }),
  unwrap: (value: unknown) => Promise.resolve(value),
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('../../shared/hooks/queries/useSignedContracts', () => ({
  useSignedContracts: () => ({ data: null }),
}));

vi.mock('../../shared/components/PageHeader', () => ({ default: () => null }));

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  createInvoiceMock.mockReset();
  updateInvoiceStatusMock.mockReset();
  listCrmCustomersMock.mockReset();
});

const renderNewInvoice = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <NewInvoice />
    </QueryClientProvider>
  );
};

const fillInvoiceToFinalStep = async () => {
  listCrmCustomersMock.mockResolvedValue({
    customers: [{ id: 42, name: 'Acme Corp', email: 'billing@acme.test' }],
  });

  renderNewInvoice();

  await screen.findByText('Acme Corp');
  fireEvent.click(screen.getByRole('button', { name: /Next/ }));
  fireEvent.change(screen.getByPlaceholderText('Description'), {
    target: { value: 'Consulting' },
  });
  const [, priceInput] = screen.getAllByRole('textbox');
  fireEvent.change(priceInput, { target: { value: '100' } });
  fireEvent.click(screen.getByRole('button', { name: /Next/ }));
  fireEvent.click(screen.getByRole('button', { name: /Next/ }));
};

describe('NewInvoice save flow', () => {
  it('creates the invoice only once while the Sent transition is still pending', async () => {
    createInvoiceMock.mockResolvedValue({
      invoice: { id: 'inv-1', clientName: 'Acme Corp', status: 'draft' },
    });
    // The status PATCH never settles: the save flow is still in flight.
    updateInvoiceStatusMock.mockReturnValue(new Promise(() => {}));

    await fillInvoiceToFinalStep();

    fireEvent.click(screen.getByRole('button', { name: 'Save & Send' }));
    await waitFor(() => expect(updateInvoiceStatusMock).toHaveBeenCalledTimes(1));

    // Both footer buttons wait for the whole flow, not just the create call.
    const savingButtons = screen.getAllByRole('button', { name: 'Saving...' });
    expect(savingButtons).toHaveLength(2);
    for (const button of savingButtons) {
      expect(button).toBeDisabled();
    }

    // The second click of a double-click landing after create resolves.
    for (const button of savingButtons) {
      fireEvent.click(button);
    }
    expect(createInvoiceMock).toHaveBeenCalledTimes(1);
  });
});
