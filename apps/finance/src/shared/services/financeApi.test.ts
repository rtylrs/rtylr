import { describe, expect, it, vi } from 'vitest';

const mockClients = vi.hoisted(() => {
  const makeClient = () => {
    const methods: Record<string, ReturnType<typeof vi.fn>> = {};
    return new Proxy(methods, {
      get(target, property: string) {
        target[property] ??= vi.fn(async () => ({}));
        return target[property];
      },
    });
  };

  return { compatibility: makeClient(), invoices: makeClient(), expenses: makeClient() };
});

vi.mock('../lib/clients', () => ({
  financeCompatibilityClient: mockClients.compatibility,
  invoicesClient: mockClients.invoices,
  expensesClient: mockClients.expenses,
}));

import { financeApi } from './financeApi';

describe('financeApi', () => {
  it('routes legacy finance paths through generated compatibility clients', async () => {
    financeApi.getDashboard();
    financeApi.getInvoices({ status: 'paid' });
    financeApi.getInvoiceDetail('INV-001');
    financeApi.getPL('2026-Q1');
    financeApi.getQuotes();
    await Promise.resolve();

    expect(mockClients.compatibility.getDashboard).toHaveBeenCalledWith({});
    expect(mockClients.compatibility.listInvoices).toHaveBeenCalledWith({ status: 'paid' });
    expect(mockClients.compatibility.getInvoice).toHaveBeenCalledWith({ id: 'INV-001' });
    expect(mockClients.compatibility.getPL).toHaveBeenCalledWith(
      expect.objectContaining({ period: '2026-Q1' })
    );
    expect(mockClients.compatibility.listQuotes).toHaveBeenCalledWith({});
  });

  it('puts the path id beside the body, not inside it, and snake_cases the body', async () => {
    financeApi.updateInvoice('INV-001', { clientEmail: 'billing@example.com' });
    financeApi.recordPayment('INV-001', 500);
    financeApi.createRecurringInvoice({
      templateInvoiceId: 4 as never,
      frequency: 'monthly',
      nextInvoiceDate: '2026-04-01',
    });
    await Promise.resolve();

    // UpdateInvoiceRequest is { id, body }: id binds the /invoices/{id} path
    // segment and cannot travel inside the payload. Burying it in `body` left
    // the URL without an id, which is what these calls used to do.
    expect(mockClients.compatibility.updateInvoice).toHaveBeenCalledWith({
      id: 'INV-001',
      body: { client_email: 'billing@example.com' },
    });

    // Creates are flat. RecordPaymentRequest and CreateRecurringInvoiceRequest
    // declare their fields directly, with no `body` member to wrap them in.
    expect(mockClients.compatibility.recordPayment).toHaveBeenCalledWith({
      invoice_id: 1,
      amount: 500,
    });
    expect(mockClients.compatibility.createRecurringInvoice).toHaveBeenCalledWith({
      template_invoice_id: 4,
      frequency: 'monthly',
      next_invoice_date: '2026-04-01',
      end_date: undefined,
      is_active: true,
    });
  });

  it('uses canonical generated delete clients where the contracts exist', async () => {
    financeApi.deleteInvoice('INV-001');
    financeApi.deleteExpense('EXP-1');
    financeApi.deleteQuote('Q-1');
    await Promise.resolve();

    expect(mockClients.invoices.deleteInvoice).toHaveBeenCalledWith({ id: 'INV-001' });
    expect(mockClients.expenses.deleteExpense).toHaveBeenCalledWith({ id: 'EXP-1' });
    expect(mockClients.compatibility.deleteQuote).toHaveBeenCalledWith({ id: 'Q-1' });
  });
});
