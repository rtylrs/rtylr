import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Bills from './Bills';

/**
 * Bulk mark-as-paid. rtylr#944.
 *
 * The reported symptom was a green success toast on a flow that always failed.
 * The cause was narrower and worse than the report: useBills maps
 * `vendor: b.description` and exposes no `description` field, so payAll's
 * `(b as Record<string, unknown>).description` was always undefined and every
 * request carried `description: ''` against a `@required` field. Nothing could
 * ever succeed, and had it succeeded it would have blanked the stored
 * description, which is where the vendor name lives.
 */

const bill = {
  id: '7',
  supplier: '42',
  supplierName: 'Acme Corp',
  vendor: 'Acme Corp invoice for March',
  description: 'Acme Corp invoice for March',
  amount: 1250,
  currency: 'EUR',
  status: 'unpaid',
  dueDate: '2026-09-30',
};

/** A one-off bill: no supplier record, so no supplier id to send. */
const oneOffBill = {
  ...bill,
  id: '8',
  supplier: '',
  supplierName: 'Corner shop',
  vendor: 'Stationery, cash',
  description: 'Stationery, cash',
};

const { useBillsMock, mutateAsyncMock, toastMock } = vi.hoisted(() => ({
  useBillsMock: vi.fn(),
  mutateAsyncMock: vi.fn(),
  toastMock: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({ toast: toastMock }));
vi.mock('@voxire/ui', () => ({ EmptyState: () => null }));

vi.mock('./useBills', () => ({ useBills: useBillsMock, BILLS_PAGE_SIZE: 50 }));
vi.mock('../../shared/hooks/queries/useBills', () => ({
  useCreateBill: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateBill: () => ({ mutateAsync: mutateAsyncMock, isPending: false }),
}));
vi.mock('../../shared/hooks/queries/useErpReceivedPurchaseOrders', () => ({
  useErpReceivedPurchaseOrders: () => ({ data: [] }),
}));
vi.mock('../../shared/lib/settings', () => ({
  useSettings: () => ({ settings: { currency: 'USD' } }),
}));
vi.mock('../../shared/components/PageHeader', () => ({
  default: ({ actions }: { actions?: React.ReactNode }) => <div>{actions}</div>,
}));
vi.mock('../../shared/components/MoneyDisplay', () => ({
  default: ({ amount }: { amount?: number }) => <span>{amount}</span>,
}));
vi.mock('../../shared/components/Skeleton', () => ({ SkeletonRow: () => null }));
vi.mock('../../shared/components/Modal', () => ({ default: () => null }));
vi.mock('../../shared/components/CurrencyInput', () => ({ default: () => null }));

/** Select every row, then press the bulk pay control. */
async function selectAllAndPay(): Promise<void> {
  const boxes = screen.getAllByRole('checkbox');
  for (const box of boxes) fireEvent.click(box);
  const payButton = await screen.findByRole('button', { name: /bills\.paySelected|Pay selected/i });
  fireEvent.click(payButton);
}

describe('Bills bulk mark-as-paid (rtylr#944)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mutateAsyncMock.mockResolvedValue({ message: 'ok' });
    useBillsMock.mockReturnValue({ data: [bill], isLoading: false });
  });

  it('round-trips the stored description instead of blanking it', async () => {
    render(<Bills />);
    await selectAllAndPay();

    await waitFor(() => expect(mutateAsyncMock).toHaveBeenCalledTimes(1));
    const sent = mutateAsyncMock.mock.calls[0][0];
    // The field is @required, and it carries the vendor name. An empty string
    // both fails validation and destroys data.
    expect(sent.body.description).toBe(bill.description);
  });

  it('keeps the stored currency rather than assuming USD', async () => {
    render(<Bills />);
    await selectAllAndPay();

    await waitFor(() => expect(mutateAsyncMock).toHaveBeenCalledTimes(1));
    expect(mutateAsyncMock.mock.calls[0][0].body.currency).toBe('EUR');
  });

  it('omits supplierId for a one-off bill rather than sending an empty string', async () => {
    useBillsMock.mockReturnValue({ data: [oneOffBill], isLoading: false });
    render(<Bills />);
    await selectAllAndPay();

    await waitFor(() => expect(mutateAsyncMock).toHaveBeenCalledTimes(1));
    const body = mutateAsyncMock.mock.calls[0][0].body;
    // The contract is explicit: "Callers must not invent a value: a bill with no
    // known supplier records none." An empty string fails the int64 regex.
    expect(body.supplierId).toBeUndefined();
    expect(Object.prototype.hasOwnProperty.call(body, 'supplierId')).toBe(false);
  });

  it('sends the supplier id when the bill has one', async () => {
    render(<Bills />);
    await selectAllAndPay();

    await waitFor(() => expect(mutateAsyncMock).toHaveBeenCalledTimes(1));
    expect(mutateAsyncMock.mock.calls[0][0].body.supplierId).toBe('42');
  });

  it('does not claim success when the update fails', async () => {
    mutateAsyncMock.mockRejectedValue(new Error('invalid request'));
    render(<Bills />);
    await selectAllAndPay();

    await waitFor(() => expect(mutateAsyncMock).toHaveBeenCalledTimes(1));
    // The defect this test exists for: a green toast on a flow that failed.
    expect(toastMock.success).not.toHaveBeenCalled();
    expect(toastMock.error).toHaveBeenCalled();
  });

  it('reports a partial result rather than a clean success', async () => {
    useBillsMock.mockReturnValue({ data: [bill, oneOffBill], isLoading: false });
    mutateAsyncMock
      .mockResolvedValueOnce({ message: 'ok' })
      .mockRejectedValueOnce(new Error('invalid request'));
    render(<Bills />);
    await selectAllAndPay();

    await waitFor(() => expect(mutateAsyncMock).toHaveBeenCalledTimes(2));
    expect(toastMock.success).not.toHaveBeenCalled();
  });

  it('claims success only when every update resolved', async () => {
    render(<Bills />);
    await selectAllAndPay();

    await waitFor(() => expect(toastMock.success).toHaveBeenCalledTimes(1));
    expect(toastMock.error).not.toHaveBeenCalled();
  });
});
